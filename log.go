package log

import (
	"context"
	"fmt"
	"io"
	"os"
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
}

var (
	global *Log = nil
)

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

	if kv.GetBool("global", false) {
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

func (l *Log) Write(lvl string, msg string) {
	if l == nil {
		return
	}

	now := l.rotate()

	l.mt.Lock()
	defer l.mt.Unlock()

	str := fmt.Sprintf("%s|%s| %s\n", now.UTC().Format(timeFormat), lvl, msg)

	if l.wh != nil {
		fmt.Fprintf(l.wh, str)
	}

	if l.stderr {
		fmt.Fprintf(os.Stderr, str)
	}

	if l.stdout {
		fmt.Println(str)
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
}

func Write(lvl string, msg string) {
	global.Write(lvl, msg)
	global = nil
}
