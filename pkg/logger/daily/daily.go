package daily

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultPattern       = "2006-01-02.log"
	defaultBufferSize    = 200
	defaultGCInterval    = time.Hour
	defaultGCFileExpired = time.Hour * 24 * 7
)

type Daily interface {
	Write([]byte) (int, error)
	Close() error
	OnGC(handler GCHandler)
	OnBeforeFileSwitch(handler BeforeFileSwitchHandler)
}

type Option func(*dailyWriter)
type GCHandler func(time.Duration, []string)
type BeforeFileSwitchHandler func(prev string, current string)

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
}

func WithPattern(format string) Option {
	return func(d *dailyWriter) {
		d.pattern = format
	}
}

func WithGCInterval(duration time.Duration) Option {
	return func(d *dailyWriter) {
		d.gcDuration = duration
	}
}

func WithGCFileExpired(expired time.Duration) Option {
	return func(d *dailyWriter) {
		d.gcFileExpired = expired
	}
}

func WithBufferSize(size int) Option {
	return func(d *dailyWriter) {
		d.bufferSize = size
	}
}

func New(folder string, options ...Option) (Daily, error) {
	err := os.MkdirAll(folder, 0755)
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

func (d *dailyWriter) Write(data []byte) (int, error) {
	select {
	case <-d.closed:
		return 0, io.ErrClosedPipe
	default:
		d.input <- data
	}

	return len(data), nil
}

func (d *dailyWriter) Close() error {
	select {
	case <-d.closed:
		return nil
	default:
		close(d.closed)
		close(d.input)
		if d.gcTicker != nil {
			d.gcTicker.Stop()
		}
	}

	<-d.done
	return nil
}

func (d *dailyWriter) OnBeforeFileSwitch(handler BeforeFileSwitchHandler) {
	d.callbackMutex.Lock()
	d.onBeforeFileSwitch = handler
	d.callbackMutex.Unlock()
}

func (d *dailyWriter) OnGC(handler GCHandler) {
	d.callbackMutex.Lock()
	d.onGC = handler
	d.callbackMutex.Unlock()
}

func (d *dailyWriter) closeOut() {
	if d.out != nil {
		d.out.Close()
	}
}

func (d *dailyWriter) dailyWorker() {
	defer close(d.done)
	defer d.closeOut()

	for data := range d.input {
		d.rotateOut()
		if d.out != nil {
			d.out.Write(data)
		}
	}
}

func (d *dailyWriter) rotateOut() {
	now := time.Now().Format(d.pattern)
	if now == d.current {
		return
	}

	d.closeOut()
	out, err := os.OpenFile(filepath.Join(d.folder, now), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
			os.Remove(filePath)
			deletes = append(deletes, fileName)
		}
	}

	d.callbackMutex.RLock()
	if d.onGC != nil {
		d.onGC(time.Since(start), deletes)
	}
	d.callbackMutex.RUnlock()
}

func (d *dailyWriter) gcWorker() {
	d.gcTicker = time.NewTicker(d.gcDuration)
	for range d.gcTicker.C {
		d.gc()
	}
}
