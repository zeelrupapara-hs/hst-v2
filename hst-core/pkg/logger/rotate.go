package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyWriter writes <dir>/YYYYMMDD.log, rolling at midnight.
type DailyWriter struct {
	dir     string
	maxAge  time.Duration
	mu      sync.Mutex
	file    *os.File
	day     string // YYYYMMDD currently open
	nowFunc func() time.Time
}

// NewDailyWriter opens <dir>. maxAgeDays 0 keeps files forever.
func NewDailyWriter(dir string, maxAgeDays int) (*DailyWriter, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	w := &DailyWriter{
		dir:     dir,
		maxAge:  time.Duration(maxAgeDays) * 24 * time.Hour,
		nowFunc: time.Now,
	}
	if err := w.rotate(w.nowFunc()); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *DailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.nowFunc()
	if day := now.Format("20060102"); day != w.day {
		if err := w.rotate(now); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

// rotate opens the file for t's date. Caller holds the lock.
func (w *DailyWriter) rotate(t time.Time) error {
	if w.file != nil {
		_ = w.file.Close()
	}

	day := t.Format("20060102")
	path := filepath.Join(w.dir, day+".log")

	f, err := os.OpenFile(filepath.Clean(path), //nosec G304 -- path built from config dir + date
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", path, err)
	}
	w.file, w.day = f, day

	// prune on rollover
	go w.prune(t)
	return nil
}

// prune deletes day files older than maxAge.
func (w *DailyWriter) prune(now time.Time) {
	if w.maxAge <= 0 {
		return
	}
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}
	cutoff := now.Add(-w.maxAge)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".log" {
			continue
		}
		day, err := time.Parse("20060102", e.Name()[:len(e.Name())-4])
		if err != nil {
			continue // not one of ours
		}
		if day.Before(cutoff) {
			_ = os.Remove(filepath.Join(w.dir, e.Name()))
		}
	}
}

func (w *DailyWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
