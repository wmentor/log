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

	for i := 0; i < 20; i++ {
		Write("info", strconv.Itoa(i+1))
	}
}
