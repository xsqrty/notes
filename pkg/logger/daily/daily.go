package daily

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// defaultPattern defines the default log file naming pattern as a date format "2006-01-02.log".
	defaultPattern = "2006-01-02.log"
	// defaultBufferSize specifies the default size of the log buffer.
	defaultBufferSize = 200
	// defaultGCInterval sets the default interval for garbage collection.
	defaultGCInterval = time.Hour
	// defaultGCFileExpired indicates the default expiration duration for log files (7 days).
	defaultGCFileExpired = time.Hour * 24 * 7
)

// Daily represents an interface for managing daily file-based logging.
type Daily interface {
	// Write writes the provided byte slice into the logging system or file. Returns the number of bytes written or an error.
	Write([]byte) (int, error)
	// Close releases all resources associated with the instance, ensuring proper cleanup and completing pending operations.
	Close() error
	// OnGC registers a callback function to handle garbage collection of expired files during regular cleanup operations.
	OnGC(handler GCHandler)
	// OnBeforeFileSwitch registers a callback executed before the file is switched during the daily writing process.
	OnBeforeFileSwitch(handler BeforeFileSwitchHandler)
}

type (
	// Option defines a functional parameter for configuring a dailyWriter instance.
	Option func(*dailyWriter)
	// GCHandler represent a function to handle garbage collection of expired files.
	GCHandler func(time.Duration, []string)
	// BeforeFileSwitchHandler defines a callback executed before file switching occurs.
	BeforeFileSwitchHandler func(prev string, current string)
)

// dailyWriter is a structure managing daily log rotation and file garbage collection efficiently.
type dailyWriter struct {
	callbackMutex      sync.RWMutex
	folder             string
	pattern            string
	current            string
	onBeforeFileSwitch BeforeFileSwitchHandler
	onGC               GCHandler
	gcTicker           *time.Ticker
	gcDuration         time.Duration
	gcFileExpired      time.Duration
	out                io.WriteCloser
	bufferSize         int
	input              chan []byte
	done               chan struct{}
	closed             chan struct{}
	closeOnce          sync.Once
}

// WithPattern sets the file naming pattern for the dailyWriter instance.
func WithPattern(format string) Option {
	return func(d *dailyWriter) {
		d.pattern = format
	}
}

// WithGCInterval sets the garbage collection interval duration for the dailyWriter.
func WithGCInterval(duration time.Duration) Option {
	return func(d *dailyWriter) {
		d.gcDuration = duration
	}
}

// WithGCFileExpired sets the duration after which old files are considered expired and removed during garbage collection.
func WithGCFileExpired(expired time.Duration) Option {
	return func(d *dailyWriter) {
		d.gcFileExpired = expired
	}
}

// WithBufferSize sets the buffer size for the dailyWriter.
func WithBufferSize(size int) Option {
	return func(d *dailyWriter) {
		d.bufferSize = size
	}
}

// New initializes and returns a Daily instance configured with the specified folder and other options.
func New(folder string, options ...Option) (Daily, error) {
	err := os.MkdirAll(folder, 0o755) // nolint: gosec
	if err != nil {
		return nil, err
	}

	daily := &dailyWriter{
		folder:        folder,
		pattern:       defaultPattern,
		gcDuration:    defaultGCInterval,
		gcFileExpired: defaultGCFileExpired,
		bufferSize:    defaultBufferSize,
		done:          make(chan struct{}),
		closed:        make(chan struct{}),
	}

	for _, option := range options {
		option(daily)
	}

	daily.input = make(chan []byte, daily.bufferSize)

	go daily.gc()
	go daily.dailyWorker()
	go daily.gcWorker()

	return daily, nil
}

// Write handles incoming data by sending it to the input channel for further processing or writing to a file.
// If the writer is closed, it returns an io.ErrClosedPipe error.
// It returns the length of the data written or an error, if any.
func (d *dailyWriter) Write(data []byte) (int, error) {
	select {
	case <-d.closed:
		return 0, io.ErrClosedPipe
	default:
		d.input <- data
	}

	return len(data), nil
}

// Close gracefully shuts down the dailyWriter, ensuring all resources are released and pending operations are completed.
func (d *dailyWriter) Close() error {
	d.closeOnce.Do(func() {
		close(d.closed)
		close(d.input)
		if d.gcTicker != nil {
			d.gcTicker.Stop()
		}
		<-d.done
	})

	return nil
}

// OnBeforeFileSwitch sets a handler to be invoked before switching to a new log file during rotation.
func (d *dailyWriter) OnBeforeFileSwitch(handler BeforeFileSwitchHandler) {
	d.callbackMutex.Lock()
	d.onBeforeFileSwitch = handler
	d.callbackMutex.Unlock()
}

// OnGC sets a handler to be executed during the garbage collection process to handle file cleanup results and duration.
func (d *dailyWriter) OnGC(handler GCHandler) {
	d.callbackMutex.Lock()
	d.onGC = handler
	d.callbackMutex.Unlock()
}

// closeOut closes the writer's current output stream if it is not nil, ensuring safe resource cleanup.
func (d *dailyWriter) closeOut() {
	if d.out != nil {
		d.out.Close() // nolint: gosec, errcheck
	}
}

// dailyWorker is a goroutine that processes data from the input channel, manages file rotation, and writes data to the output.
func (d *dailyWriter) dailyWorker() {
	defer close(d.done)
	defer d.closeOut()

	for data := range d.input {
		d.rotateOut()
		if d.out != nil {
			d.out.Write(data) // nolint: gosec, errcheck
		}
	}
}

// rotateOut handles the rotation of the output file based on time, ensuring that data is written to the correct file.
func (d *dailyWriter) rotateOut() {
	now := time.Now().Format(d.pattern)
	if now == d.current {
		return
	}

	d.closeOut()
	out, err := os.OpenFile(filepath.Join(d.folder, now), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644) // nolint: gosec
	if err != nil {
		panic(err)
	}

	d.callbackMutex.RLock()
	if d.onBeforeFileSwitch != nil && d.current != "" {
		d.onBeforeFileSwitch(d.current, now)
	}
	d.callbackMutex.RUnlock()

	d.current = now
	d.out = out
}

// gc performs garbage collection by removing expired log files from the specified folder based on the configured expiration time.
func (d *dailyWriter) gc() {
	dir, err := os.ReadDir(d.folder)
	if err != nil {
		return
	}

	start := time.Now()
	deletes := make([]string, 0)
	for _, fi := range dir {
		if !fi.Type().IsRegular() {
			continue
		}

		fileName := fi.Name()
		filePath := filepath.Join(d.folder, fileName)
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if start.Sub(info.ModTime()) > d.gcFileExpired {
			if err := os.Remove(filePath); err != nil {
				continue
			}

			deletes = append(deletes, fileName)
		}
	}

	d.callbackMutex.RLock()
	if d.onGC != nil {
		d.onGC(time.Since(start), deletes)
	}
	d.callbackMutex.RUnlock()
}

// gcWorker continuously triggers garbage collection at regular intervals defined by gcDuration using a ticker.
func (d *dailyWriter) gcWorker() {
	d.gcTicker = time.NewTicker(d.gcDuration)
	for range d.gcTicker.C {
		d.gc()
	}
}
