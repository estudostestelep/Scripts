package main

import (
	"fmt"
	"time"
)

type Logger struct {
	verbose bool
}

const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorGreen   = "\033[32m"
	colorRed     = "\033[31m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorCyan    = "\033[36m"
)

func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("%s[ℹ]%s %s\n", colorBlue, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Success(format string, args ...interface{}) {
	fmt.Printf("%s[✓]%s %s\n", colorGreen, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Printf("%s[✗]%s %s\n", colorRed, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(format string, args ...interface{}) {
	fmt.Printf("%s[⚠]%s %s\n", colorYellow, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Skip(format string, args ...interface{}) {
	fmt.Printf("%s[⏭]%s %s\n", colorCyan, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Printf("%s[D]%s %s\n", colorYellow, colorReset, fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Section(title string) {
	fmt.Printf("\n%s%s========== %s ==========%s\n", colorBold, colorBlue, title, colorReset)
}

func (l *Logger) Subsection(title string) {
	fmt.Printf("\n%s>>> %s%s\n", colorCyan, title, colorReset)
}

func (l *Logger) Summary(title string, created, skipped, failed int) {
	fmt.Printf("\n%s%s========== %s ==========%s\n", colorBold, colorBlue, title, colorReset)
	fmt.Printf("%s[✓]%s Criados: %d\n", colorGreen, colorReset, created)
	fmt.Printf("%s[⏭]%s Já existiam: %d\n", colorCyan, colorReset, skipped)
	if failed > 0 {
		fmt.Printf("%s[✗]%s Erros: %d\n", colorRed, colorReset, failed)
	}
	fmt.Printf("%s========== Fim: %s ==========%s\n\n", colorBlue, time.Now().Format("15:04:05"), colorReset)
}
