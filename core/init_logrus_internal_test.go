package core

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestLogFormatterFormat(t *testing.T) {
	f := &LogFormatter{}
	logger := logrus.New()
	logger.SetReportCaller(true)

	entry := &logrus.Entry{
		Logger:  logger,
		Time:    time.Date(2026, 3, 2, 12, 0, 0, 0, time.Local),
		Level:   logrus.InfoLevel,
		Message: "hello",
	}

	b, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format 失败: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "hello") || !strings.Contains(s, "[info]") {
		t.Fatalf("无 caller 格式输出异常: %s", s)
	}

	entry.Caller = &runtime.Frame{
		File:     "a/b/c.go",
		Line:     12,
		Function: "pkg.fn",
	}
	b, err = f.Format(entry)
	if err != nil {
		t.Fatalf("带 caller 的 Format 失败: %v", err)
	}
	s = string(b)
	if !strings.Contains(s, "c.go:12") || !strings.Contains(s, "pkg.fn") {
		t.Fatalf("带 caller 格式输出异常: %s", s)
	}
}

func TestFileDateHookFire(t *testing.T) {
	dir := t.TempDir()
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableColors: true})

	hook := &FileDateHook{
		logPath:  dir,
		fileDate: "2000-01-01",
		appName:  "app",
	}

	entry := &logrus.Entry{
		Logger:  logger,
		Time:    time.Date(2026, 3, 2, 9, 0, 0, 0, time.Local),
		Level:   logrus.InfoLevel,
		Message: "line-1",
		Buffer:  &bytes.Buffer{},
	}

	if err := hook.Fire(entry); err != nil {
		t.Fatalf("首次 Fire 失败: %v", err)
	}
	if hook.file == nil {
		t.Fatal("首次 Fire 后应创建日志文件")
	}
	t.Cleanup(func() {
		_ = hook.file.Close()
	})

	entry2 := &logrus.Entry{
		Logger:  logger,
		Time:    time.Date(2026, 3, 2, 10, 0, 0, 0, time.Local),
		Level:   logrus.WarnLevel,
		Message: "line-2",
		Buffer:  &bytes.Buffer{},
	}
	if err := hook.Fire(entry2); err != nil {
		t.Fatalf("同日再次 Fire 失败: %v", err)
	}

	logFile := filepath.Join(dir, "2026-03-02", "app.log")
	b, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "line-1") || !strings.Contains(s, "line-2") {
		t.Fatalf("日志内容不完整: %s", s)
	}
}
