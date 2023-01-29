package main

import "sync"

type Logger struct {
	lines []string
	mx    sync.RWMutex
	n     int
	cb    func()
}

func NewLogger() *Logger {
	return &Logger{
		lines: make([]string, 0),
		mx:    sync.RWMutex{},
		n:     100,
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.AddLine(string(p))
	n = len(p)
	return
}

func (l *Logger) AddLine(s string) {
	l.mx.Lock()

	l.lines = append(l.lines, s)
	if len(l.lines) > l.n {
		l.lines = l.lines[len(l.lines)-l.n:]
	}
	l.mx.Unlock()
	if l.cb != nil {
		l.cb()
	}
}

func (l *Logger) GetLines(n int) []string {
	l.mx.RLock()
	defer l.mx.RUnlock()
	if len(l.lines) <= n {
		return l.lines[:]
	}
	return l.lines[len(l.lines)-n:]
}
