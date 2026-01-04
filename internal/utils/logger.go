package utils

import (
	"fmt"
	"time"
)

type Logger interface {
	SetLevel(debug bool)
	LogError(err error)
	LogNewError(format string, args ...any)
	LogInfo(format string, args ...any)
	Log(format string, args ...any)
	LogCustom(color string, context string, msg string)
}

type logger struct {
	Level    int
	ShowTime bool
}

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[92m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

// Tag definitions for logger
const (
	ErrorTAG = 1
	InfoTAG  = 2
	LogTAG   = 3
)

var tagToString = map[int]string{
	ErrorTAG: Red + "[ERROR]" + Reset,
	InfoTAG:  Blue + "[INFO]" + Reset,
	LogTAG:   Cyan + "[LOG]" + Reset,
}

const (
	DebugLevel = 1
	ProdLevel  = 2
)

func NewLogger(LogLevel int, Showtime bool) Logger {
	return &logger{
		Level:    LogLevel,
		ShowTime: Showtime,
	}
}

func getTime() string {
	return fmt.Sprintf("%v", time.Now().Format("2006-01-02 15:04:05"))
}

func (l *logger) SetLevel(debug bool) {
	if debug {
		l.Level = DebugLevel
	} else {
		l.Level = ProdLevel
	}
}

// Logs some information with the provided tag and message
func (l *logger) out(tag int, msg string) {
	s := tagToString[tag] + White + ": " + msg + Reset
	if l.ShowTime {
		s = getTime() + " " + s
	}
	fmt.Println(s)
}

// Logs an error message
func (l *logger) LogError(err error) {
	if err == nil {
		return
	}
	l.out(ErrorTAG, err.Error())
}

// Logs a custom error message
func (l *logger) LogNewError(format string, args ...any) {
	l.out(ErrorTAG, fmt.Sprintf(format, args...))
}

// Logs a debug/info message if debug mode is enabled
func (l *logger) LogInfo(format string, args ...any) {
	l.out(InfoTAG, fmt.Sprintf(format, args...))
}

func (l *logger) Log(format string, args ...any) {
	if l.Level == DebugLevel {
		l.out(LogTAG, fmt.Sprintf(format, args...))
	}
}

func (l *logger) LogCustom(color string, context string, msg string) {
	fmt.Println(getTime(), color+"["+context+"]"+Reset+White+": "+msg+Reset)
}
