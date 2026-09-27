package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	mu    sync.Mutex
	out   io.Writer
	level Level
}

var std = &Logger{out: os.Stdout, level: LevelInfo}

var (
	ringBuffer []string
	ringMu     sync.RWMutex
	listeners  = make(map[chan string]struct{})
	listenerMu sync.Mutex
	maxBuffer  = 1000
)

func AddListener(ch chan string) {
	listenerMu.Lock()
	listeners[ch] = struct{}{}
	listenerMu.Unlock()
}

func RemoveListener(ch chan string) {
	listenerMu.Lock()
	delete(listeners, ch)
	listenerMu.Unlock()
}

func GetBufferedLogs() []string {
	ringMu.RLock()
	defer ringMu.RUnlock()
	res := make([]string, len(ringBuffer))
	copy(res, ringBuffer)
	return res
}

func ClearLogs() {
	ringMu.Lock()
	ringBuffer = nil
	ringMu.Unlock()

	listenerMu.Lock()
	defer listenerMu.Unlock()
	for ch := range listeners {
		select {
		case ch <- "__CLEAR__":
		default:
		}
	}
}

func SetLevel(l Level)      { std.level = l }
func SetOutput(w io.Writer) { std.out = w }

func Debug(tag, msg string, args ...any) { std.log(LevelDebug, tag, msg, args...) }
func Info(tag, msg string, args ...any)  { std.log(LevelInfo, tag, msg, args...) }
func Warn(tag, msg string, args ...any)  { std.log(LevelWarn, tag, msg, args...) }
func Error(tag, msg string, args ...any) { std.log(LevelError, tag, msg, args...) }

func (l *Logger) log(lvl Level, tag, msg string, args ...any) {
	if lvl < l.level {
		return
	}
	label := levelLabel(lvl)
	ts := time.Now().Format("2006-01-02 15:04:05")
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(label)
	sb.WriteString("] ")
	sb.WriteString(ts)
	sb.WriteString(" ")
	sb.WriteString(tag)
	sb.WriteString(": ")
	sb.WriteString(msg)
	if len(args) > 0 {
		for i := 0; i+1 < len(args); i += 2 {
			sb.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
		}
	}
	line := sb.String()

	// 1. Write to stdout
	l.mu.Lock()
	l.out.Write([]byte(line + "\n"))
	l.mu.Unlock()

	// 2. Append to ring buffer
	ringMu.Lock()
	ringBuffer = append(ringBuffer, line)
	if len(ringBuffer) > maxBuffer {
		ringBuffer = ringBuffer[len(ringBuffer)-maxBuffer:]
	}
	ringMu.Unlock()

	// 3. Broadcast to SSE listeners
	listenerMu.Lock()
	for ch := range listeners {
		select {
		case ch <- line:
		default:
		}
	}
	listenerMu.Unlock()
}

func levelLabel(l Level) string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	}
	return "INFO"
}

func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
