package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/pkg/config/formatter"
	"github.com/xsqrty/notes/pkg/logger/daily"
)

// Logger is a wrapper around zerolog.Logger that adds support for daily file-based logging and lifecycle management.
type Logger struct {
	zerolog.Logger
	daily daily.Daily
}

// NewLogger initializes and returns a new Logger instance based on the provided LoggerConfig settings.
// It supports multiple output targets like stdout and file-based logging with options for JSON or Pretty formatting.
// Returns an error if configuration is invalid or logging setup fails.
func NewLogger(loggerConfig config.LoggerConfig) (*Logger, error) {
	var log Logger
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	level, err := zerolog.ParseLevel(loggerConfig.Level)
	if err != nil {
		return nil, err
	}

	var outputs []io.Writer
	if loggerConfig.Stdout && loggerConfig.StdoutFormater == formatter.Json {
		outputs = append(outputs, os.Stdout)
	} else if loggerConfig.Stdout && loggerConfig.StdoutFormater == formatter.Pretty {
		outputs = append(outputs, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05.000",
			FormatLevel: func(i interface{}) string {
				level := fmt.Sprintf("[%s]", strings.ToUpper(i.(string)))
				return level
			},
		})
	}

	if loggerConfig.FileOut && loggerConfig.DirOut != "" {
		d, err := daily.New(loggerConfig.DirOut)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, d)
		log.daily = d
	}

	if len(outputs) == 0 {
		return nil, errors.New("log writers is not configured")
	}

	log.Logger = zerolog.New(io.MultiWriter(outputs...)).Level(level).With().Timestamp().Logger()
	if log.daily != nil {
		log.daily.OnBeforeFileSwitch(func(prev string, current string) {
			log.Info().Str("from", prev).Str("to", current).Msg("Stream switched")
		})

		log.daily.OnGC(func(duration time.Duration, files []string) {
			log.Info().
				Str("completed_at", duration.String()).
				Strs("files", files).
				Msg(fmt.Sprintf("GC completed, %d files deleted", len(files)))
		})
	}

	return &log, nil
}

// Close releases resources associated with the Logger, including closing any underlying daily file if applicable.
func (l *Logger) Close() error {
	if l.daily == nil {
		return nil
	}

	return l.daily.Close()
}
