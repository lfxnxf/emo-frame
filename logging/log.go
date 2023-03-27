package logging

import (
	"github.com/lfxnxf/emo-frame/rolling"
	"os"
)

func NewLogging(path string) *Logger {
	l, _ := NewJSON(path, rolling.HourlyRolling)
	l.SetFlags(0)
	l.SetPrintLevel(false)
	l.SetHighlighting(false)
	l.SetOutputByName(path)
	l.SetTimeFmt(TIMEMICRO)
	return l
}

func Stdout() *Logger {
	s := New()
	s.SetFlags(1)
	s.SetPrintLevel(false)
	s.SetHighlighting(false)
	s.SetPrintTime(true)
	s.SetTimeFmt(TIMEMICRO)
	s.SetOutput(os.Stdout)
	return s
}

func Noop() *Logger {
	n := New()
	n.SetFlags(0)
	n.SetPrintLevel(false)
	n.SetHighlighting(false)
	n.SetTimeFmt(TIMEMICRO)
	out, _ := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, 0600)
	n.SetOutput(out)
	return n
}
