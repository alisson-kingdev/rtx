package utils

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red          = color.New(color.FgRed).SprintFunc()
	green        = color.New(color.FgGreen).SprintFunc()
	yellow       = color.New(color.FgYellow).SprintFunc()
	blue         = color.New(color.FgBlue).SprintFunc()
	cyan         = color.New(color.FgCyan).SprintFunc()
	magenta      = color.New(color.FgMagenta).SprintFunc()
	bold         = color.New(color.Bold).SprintFunc()
	CyanColor    = color.New(color.FgCyan)
	MagentaColor = color.New(color.FgMagenta)
	YellowColor  = color.New(color.FgYellow)
	GreenColor   = color.New(color.FgGreen)
	RedColor     = color.New(color.FgRed)
	BlueColor    = color.New(color.FgBlue)
	WhiteColor   = color.New(color.FgWhite)
)

func Colorize(color *color.Color, text string) string {
	return color.Sprint(text)
}

func Section(title string) string {
	return "\n" + yellow("[*] "+title) + "\n"
}

func Success(msg string) string {
	return green("[+] " + msg)
}

func Error(msg string) string {
	return red("[!] " + msg)
}

func Info(msg string) string {
	return blue("[*] " + msg)
}

func Warning(msg string) string {
	return magenta("[*] " + msg)
}

func PrintKV(key, value string) string {
	return fmt.Sprintf("%s%s: %s%s\n", yellow(key+":"), cyan(" "), value, "\n")
}

// Logger provides structured logging with colors
type Logger struct{}

// Log logs a message with a given level and color
func (l *Logger) Log(level, message string) {
	var prefix string
	switch level {
	case "success", "ok", "+":
		prefix = green("[+] ")
	case "error", "err", "!":
		prefix = red("[!] ")
	case "info", "*":
		prefix = blue("[*] ")
	case "warning", "warn":
		prefix = yellow("[!] ")
	default:
		prefix = blue("[*] ")
	}
	fmt.Printf("%s%s\n", prefix, message)
}

// Success logs a success message
func (l *Logger) Success(msg string) {
	l.Log("success", msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.Log("error", msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.Log("info", msg)
}

// Warning logs a warning message
func (l *Logger) Warning(msg string) {
	l.Log("warning", msg)
}

// Log is a package-level logger instance
var Log = &Logger{}
