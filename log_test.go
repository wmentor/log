package log

import (
	"strconv"
	"testing"
)

func TestLog(t *testing.T) {

	_, err := Open("path=. period=minute global=1")
	if err != nil {
		t.Fatal("Open failed")
	}
	defer Close()

	for i := 0; i < 5; i++ {
		Trace(strconv.Itoa(i + 1))
		Debug(strconv.Itoa(i + 1))
		Info(strconv.Itoa(i + 1))
		Warn(strconv.Itoa(i + 1))
		Error(strconv.Itoa(i + 1))
                Stack("error")

		Tracef("%d", i+1)
		Debugf("%d", i+1)
		Infof("%d", i+1)
		Warnf("%d", i+1)
		Errorf("%d", i+1)
	}
}
