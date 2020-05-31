package log

import (
	"strconv"
	"testing"
	"time"
)

func TestLog(t *testing.T) {

	l, err := New("path=. period=minute")
	if err != nil {
		t.Fatal("New failed")
	}
	defer l.Close()

	for i := 0; i < 20; i++ {
		l.Log("info", strconv.Itoa(i+1))
	}
}
