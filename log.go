package log

import (
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
	period string
	pf     periodFunc
	stderr bool
	stdout bool
	name   string
	path   string
	wh     io.WriteCloser
	mt     sync.Mutex
}

func periodMinute(t time.Time) string {
	return fmt.Sprintf("%04d%02d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute())
}

func periodHour(t time.Time) string {
	return fmt.Sprintf("%04d%02d%02d%02d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func periodDay(t time.Time) string {
	return fmt.Sprintf("%04d%02d%02d", t.Year(), int(t.Month()), t.Day())
}

func periodMonth(t time.Time) string {
	return fmt.Sprintf("%04d%02d", t.Year(), int(t.Month()))
}

func New(opts string) (*Log, error) {

	kv, err := dsn.New(opts)
	if err != nil {
		return nil, err
	}

	var pf periodFunc

	switch kv.GetString("period", "day") {

	case "minute":
		pf = periodMinute

	case "hour":
		pf = periodHour

	case "day":
		pf = periodDay

	case "month":
		pf = periodMonth

	default:
		pf = periodDay
	}

	l := &Log{
		period: "",
		pf:     pf,
		stderr: kv.GetBool("stderr", false),
		stdout: kv.GetBool("stdout", false),
		name:   kv.GetString("name", ""),
		path:   kv.GetString("path", "."),
	}

	l.rotate()

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

		if l.wh, err = os.OpenFile(l.makeFilename(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755); err != nil {
			return now
		}
	}

	return now
}

func (l *Log) Log(lvl string, msg string) {
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

func (l *Log) makeFilename() string {

	if l.name != "" {
		return strings.TrimRight(l.path, "/") + "/" + l.name + "-" + l.period + ".log"
	}

	return ""
}

func (l *Log) Close() {
	if l.wh != nil {
		l.wh.Close()
		l.wh = nil
	}
}
