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

func SetLevel(l Level) { std.level = l }
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
	ts := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	var sb strings.Builder
	sb.WriteString(ts)
	sb.WriteByte(' ')
	sb.WriteString(label)
	sb.WriteByte(' ')
	sb.WriteString(tag)
	sb.WriteString(": ")
	sb.WriteString(msg)
	if len(args) > 0 {
		for i := 0; i+1 < len(args); i += 2 {
			sb.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
		}
	}
	sb.WriteByte('\n')
	l.mu.Lock()
	l.out.Write([]byte(sb.String()))
	l.mu.Unlock()
}

func levelLabel(l Level) string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelWarn:
		return "WARN "
	case LevelError:
		return "ERROR"
	}
	return "INFO "
}

func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
