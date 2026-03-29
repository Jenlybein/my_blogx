package image_ref_river_service

import (
	"context"
	"fmt"
	"log/slog"
	"myblogx/utils/logsafe"
	"runtime"

	"github.com/sirupsen/logrus"
)

func logrusToSlogAdapter(logger *logrus.Logger) *slog.Logger {
	return slog.New(&simpleLogrusHandler{logger: logger})
}

type simpleLogrusHandler struct {
	logger *logrus.Logger
}

func (h *simpleLogrusHandler) Handle(_ context.Context, r slog.Record) error {
	entry := h.logger.WithFields(logrus.Fields{})
	r.Attrs(func(a slog.Attr) bool {
		key, value, ok := logsafe.SlogAttrToField(a)
		if ok {
			entry = entry.WithField(key, value)
		}
		return true
	})

	message := r.Message
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		if frame, _ := fs.Next(); frame.File != "" {
			fileName := frame.File
			for i := len(fileName) - 1; i >= 0; i-- {
				if fileName[i] == '/' || fileName[i] == '\\' {
					fileName = fileName[i+1:]
					break
				}
			}
			funcName := frame.Function
			for i := len(funcName) - 1; i >= 0; i-- {
				if funcName[i] == '.' {
					funcName = funcName[i+1:]
					break
				}
			}
			message = fmt.Sprintf("[%s:%d %s] %s", fileName, frame.Line, funcName, message)
		}
	}

	switch r.Level {
	case slog.LevelDebug:
		entry.Debug(message)
	case slog.LevelInfo:
		entry.Info(message)
	case slog.LevelWarn:
		entry.Warn(message)
	case slog.LevelError:
		entry.Error(message)
	default:
		entry.Info(message)
	}
	return nil
}

func (h *simpleLogrusHandler) WithAttrs([]slog.Attr) slog.Handler       { return h }
func (h *simpleLogrusHandler) WithGroup(string) slog.Handler            { return h }
func (h *simpleLogrusHandler) Enabled(context.Context, slog.Level) bool { return true }
