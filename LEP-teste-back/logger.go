package main

import (
	"fmt"
	"time"
)

// Color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[90m"
)

type Logger struct {
	Verbose bool
}

func NewLogger(verbose bool) *Logger {
	return &Logger{Verbose: verbose}
}

func (l *Logger) timestamp() string {
	return fmt.Sprintf("[%s]", time.Now().Format("15:04:05.000"))
}

func (l *Logger) Info(msg string, args ...interface{}) {
	fmt.Printf("%s %s%s%s\n", l.timestamp(), ColorBlue, fmt.Sprintf(msg, args...), ColorReset)
}

func (l *Logger) Success(msg string, args ...interface{}) {
	fmt.Printf("%s %s✓ %s%s\n", l.timestamp(), ColorGreen, fmt.Sprintf(msg, args...), ColorReset)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	fmt.Printf("%s %s✗ %s%s\n", l.timestamp(), ColorRed, fmt.Sprintf(msg, args...), ColorReset)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	fmt.Printf("%s %s⚠ %s%s\n", l.timestamp(), ColorYellow, fmt.Sprintf(msg, args...), ColorReset)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.Verbose {
		fmt.Printf("%s %s[DEBUG] %s%s\n", l.timestamp(), ColorGray, fmt.Sprintf(msg, args...), ColorReset)
	}
}

func (l *Logger) Section(title string) {
	fmt.Printf("\n%s%s═══════════════════════════════════════════════════════════════%s\n", ColorMagenta, ColorCyan, ColorReset)
	fmt.Printf("%s%s  %s%s\n", ColorMagenta, ColorCyan, title, ColorReset)
	fmt.Printf("%s%s═══════════════════════════════════════════════════════════════%s\n\n", ColorMagenta, ColorCyan, ColorReset)
}

func (l *Logger) Subsection(title string) {
	fmt.Printf("\n%s▸ %s%s\n", ColorCyan, title, ColorReset)
}

func (l *Logger) Stats(total, passed, failed int) {
	fmt.Printf("\n%s=== RESUMO ===%s\n", ColorCyan, ColorReset)
	fmt.Printf("Total:  %d\n", total)
	fmt.Printf("%s%s✓ Sucesso: %d%s\n", ColorGreen, "", passed, ColorReset)
	fmt.Printf("%s%s✗ Falhas:  %d%s\n", ColorRed, "", failed, ColorReset)

	if failed == 0 {
		fmt.Printf("\n%s🎉 TODOS OS TESTES PASSARAM! 🎉%s\n\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("\n%s⚠️ ALGUNS TESTES FALHARAM ⚠️%s\n\n", ColorRed, ColorReset)
	}
}
