package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ANSI color codes for terminal output
var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

// Logger handles file & console logging with rotation and cleanup
type Logger struct {
	appName   string
	logDir    string
	files     map[string]*os.File
	writers   map[string]*log.Logger
	mu        sync.Mutex
	retention time.Duration
}

// New creates a new logger with daily rotation and cleanup
//
// Example: log, _ := logger.New("mybot", "./logs", 7)
func New(appName, logDir string, retentionDays int) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	l := &Logger{
		appName:   appName,
		logDir:    logDir,
		files:     make(map[string]*os.File),
		writers:   make(map[string]*log.Logger),
		retention: time.Duration(retentionDays) * 24 * time.Hour,
	}

	if err := l.rotateLogs(); err != nil {
		return nil, err
	}

	// Run background tasks
	go l.autoRotateDaily()
	go l.cleanupOldLogs()

	return l, nil
}

// rotateLogs creates new log files for info, warn, error
func (l *Logger) rotateLogs() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close existing files if open
	for _, f := range l.files {
		f.Close()
	}

	date := time.Now().Format("2006-01-02")
	levels := []string{"info", "warn", "error"}

	for _, level := range levels {
		filename := fmt.Sprintf("log_%s_%s_%s.log", l.appName, level, date)
		fullPath := filepath.Join(l.logDir, filename)

		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s log file: %w", level, err)
		}

		// Log to both file and stdout
		writer := io.MultiWriter(os.Stdout, f)
		l.files[level] = f
		l.writers[level] = log.New(writer, "", log.LstdFlags|log.Lmicroseconds)
	}
	return nil
}

// autoRotateDaily rotates logs at midnight automatically
func (l *Logger) autoRotateDaily() {
	currentDate := time.Now().Day()
	for {
		time.Sleep(time.Hour)
		if time.Now().Day() != currentDate {
			currentDate = time.Now().Day()
			_ = l.rotateLogs()
		}
	}
}

// cleanupOldLogs deletes log files older than retention period
func (l *Logger) cleanupOldLogs() {
	for {
		files, _ := os.ReadDir(l.logDir)
		cutoff := time.Now().Add(-l.retention)

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			path := filepath.Join(l.logDir, f.Name())
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				_ = os.Remove(path)
			}
		}

		time.Sleep(24 * time.Hour)
	}
}

// Info logs informational messages
func (l *Logger) Info(format string, v ...interface{}) {
	l.output("info", colorGreen, format, v...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	l.output("warn", colorYellow, format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.output("error", colorRed, format, v...)
}

// output writes to both console and respective file
func (l *Logger) output(level, color, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	colored := fmt.Sprintf("[%s] %s%s%s", strings.ToUpper(level), color, msg, colorReset)

	writer, ok := l.writers[level]
	if !ok {
		fmt.Printf("unknown log level: %s\n", level)
		return
	}

	writer.Println(colored)
}

// Close closes all open log files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range l.files {
		f.Close()
	}
	l.files = make(map[string]*os.File)
	return nil
}
