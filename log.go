package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wmentor/dsn"
)

const (
	timeFormat string = "2006-01-02 15:04:05"
)

type periodFunc func(time.Time) string

type Log struct {
	period  string
	pf      periodFunc
	stderr  bool
	stdout  bool
	name    string
	path    string
	wh      io.WriteCloser
	mt      sync.Mutex
	rmDelta time.Duration
	cnFunc  context.CancelFunc
	level   int
}

var (
	global *Log = nil
	levels map[string]int
)

func init() {

	levels = make(map[string]int)

	for i, l := range []string{"trace", "debug", "info", "warn", "error", "fatal"} {
		levels[l] = i + 1
	}

}

func periodMinute(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%04d%02d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute())
}

func periodHour(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%04d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func periodDay(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%04d%02d%02d", t.Year(), int(t.Month()), t.Day())
}

func periodMonth(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%04d%02d", t.Year(), int(t.Month()))
}

func Open(opts string) (*Log, error) {

	kv, err := dsn.New(opts)
	if err != nil {
		return nil, err
	}

	keep := kv.GetInt("keep", 15)

	var pf periodFunc
	var delInt time.Duration

	switch kv.GetString("period", "day") {

	case "minute":
		pf = periodMinute
		delInt = time.Minute * time.Duration(keep)

	case "hour":
		pf = periodHour
		delInt = time.Hour * time.Duration(keep)

	case "day":
		pf = periodDay
		delInt = time.Hour * 24 * time.Duration(keep)

	case "month":
		pf = periodMonth
		delInt = time.Hour * 24 * 30 * time.Duration(keep)

	default:
		pf = periodDay
		delInt = time.Hour * 24 * time.Duration(keep)
	}

	l := &Log{
		period:  "",
		pf:      pf,
		stderr:  kv.GetBool("stderr", false),
		stdout:  kv.GetBool("stdout", false),
		name:    kv.GetString("name", ""),
		path:    kv.GetString("path", "."),
		rmDelta: delInt,
		level:   levels[kv.GetString("level", "info")],
	}

	if l.name != "" {
		l.rotate()

		ctx, cancel := context.WithCancel(context.Background())

		l.cnFunc = cancel

		go func(ctx context.Context, l *Log) {

			for {

				select {

				case <-ctx.Done():
					return

				case <-time.After(time.Second * 10):
					l.rotate()

				}

			}

		}(ctx, l)
	}

	if kv.GetBool("global", true) {
		global = l
	}

	return l, nil
}

func (l *Log) rotate() time.Time {

	l.mt.Lock()
	defer l.mt.Unlock()

	now := time.Now()

	if l.name == "" {
		return now
	}

	np := l.pf(now)

	if l.period != np {
		if l.wh != nil {
			l.wh.Close()
			l.wh = nil
		}

		l.period = np

		var err error

		if l.wh, err = os.OpenFile(l.makeFilename(np), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755); err != nil {
			return now
		}

		cleanPeriod := l.pf(now.Add(-l.rmDelta))
		os.Remove(l.makeFilename(cleanPeriod))
	}

	return now
}

func (l *Log) write(lvl string, msg string) {
	if l == nil {
		if lvl == "fatal" {
			os.Exit(1)
		}
		return
	}

	if l.level == 0 {
		if lvl == "fatal" {
			os.Exit(1)
		}
		return
	}

	if levels[lvl] < l.level {
		return
	}

	now := l.rotate()

	l.mt.Lock()
	defer l.mt.Unlock()

	str := fmt.Sprintf("%s|%5s| %s\n", now.UTC().Format(timeFormat), lvl, msg)

	if l.wh != nil {
		fmt.Fprint(l.wh, str)
	}

	if l.stderr {
		fmt.Fprint(os.Stderr, str)
	}

	if l.stdout {
		fmt.Print(str)
	}

	if lvl == "fatal" {
		os.Exit(1)
	}
}

func (l *Log) makeFilename(period string) string {

	if l.name != "" {
		return strings.TrimRight(l.path, "/") + "/" + l.name + "-" + period + ".log"
	}

	return ""
}

func (l *Log) Close() {

	if l == nil {
		return
	}

	if l.cnFunc != nil {
		l.cnFunc()
		l.cnFunc = nil
	}

	if l.wh != nil {
		l.wh.Close()
		l.wh = nil
	}
}

func Close() {
	global.Close()
	global = nil
}

func write(lvl string, msg string) {
	global.write(lvl, msg)
}

func (l *Log) Trace(msg string) {
	l.write("trace", msg)
}

func (l *Log) Debug(msg string) {
	l.write("debug", msg)
}

func (l *Log) Info(msg string) {
	l.write("info", msg)
}

func (l *Log) Warn(msg string) {
	l.write("warn", msg)
}

func (l *Log) Error(msg string) {
	l.write("error", msg)
}

func (l *Log) Fatal(msg string) {
	l.write("fatal", msg)
}

func (l *Log) Tracef(format string, args ...interface{}) {
	l.Trace(fmt.Sprintf(format, args...))
}

func (l *Log) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Log) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Log) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Log) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Log) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *Log) Stack(lvl string) {
	buf := make([]byte, 1024*100)
	n := runtime.Stack(buf, false)
	l.write(lvl, string(buf[:n]))
}

func Trace(msg string) {
	write("trace", msg)
}

func Debug(msg string) {
	write("debug", msg)
}

func Info(msg string) {
	write("info", msg)
}

func Warn(msg string) {
	write("warn", msg)
}

func Error(msg string) {
	write("error", msg)
}

func Fatal(msg string) {
	write("fatal", msg)
}

func Tracef(format string, args ...interface{}) {
	Trace(fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}

func Stack(lvl string) {
	global.Stack(lvl)
}
