package config

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

func (cfg *Config) Log() *Logger {
	return cfg.log
}

func (cfg *Config) SetLog(log *Logger) {
	cfg.log = log
}

type Logger struct {
	*zerolog.Logger
	panicOnFatal bool
	buffer       *LoggingBuffer
}

type LoggingBuffer struct {
	mutex sync.Mutex
	bytes.Buffer
}

func NewLogger(log *zerolog.Logger) *Logger {
	return &Logger{
		Logger: log,
	}
}

func NewTestLogger(tl zerolog.TestingLog, level zerolog.Level) *Logger {
	tl.Helper()
	buffer := &LoggingBuffer{}
	log := zerolog.New(io.MultiWriter(zerolog.NewTestWriter(tl), buffer)).Level(level)
	return &Logger{
		Logger:       &log,
		panicOnFatal: true,
		buffer:       buffer,
	}
}

func (log *Logger) Buffer() *LoggingBuffer {
	return log.buffer
}

func (log *Logger) Fatal() *zerolog.Event {
	if log.panicOnFatal {
		return log.Logger.Panic() //nolint:zerologlint
	}
	return log.Logger.Fatal() //nolint:zerologlint
}

func (b *LoggingBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.Buffer.Write(p) //nolint:wrapcheck
}

func (b *LoggingBuffer) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Buffer.Reset()
}

func (b *LoggingBuffer) Events() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, line := range strings.Split(b.String(), "\n") {
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			result = append(result, event)
		}
	}
	return result
}
