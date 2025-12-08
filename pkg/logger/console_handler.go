package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
)

type ConsoleHandler struct {
	opts  slog.HandlerOptions
	w     io.Writer
	mu    sync.Mutex
	attrs []slog.Attr
	group string
}

func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *ConsoleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	return &ConsoleHandler{
		opts: *opts,
		w:    w,
	}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// 1. Time
	if !r.Time.IsZero() {
		buf.WriteString(r.Time.Format("2006-01-02T15:04:05.999Z"))
	}
	buf.WriteByte('\t')

	// 2. Level
	buf.WriteString(r.Level.String())
	buf.WriteByte('\t')

	// 3. Source
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf.WriteString(fmt.Sprintf("%s:%d", f.File, f.Line))
	}
	buf.WriteByte('\t')

	// 4. Message
	buf.WriteString(fmt.Sprintf("%q", r.Message))

	// 5. Attributes
	var attrsBuilder strings.Builder
	allAttrs := h.attrs
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

	for _, a := range allAttrs {
		attrsBuilder.WriteByte('\t')
		attrsBuilder.WriteString(a.Value.String())
	}

	if attrsBuilder.Len() > 0 {
		buf.WriteString(attrsBuilder.String())
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	// Create a new handler; do not copy the mutex.
	h2 := &ConsoleHandler{
		opts:  h.opts,
		w:     h.w,
		group: h.group,
		// The new mutex gets its zero value (unlocked), which is correct.
		attrs: make([]slog.Attr, len(h.attrs)+len(attrs)),
	}
	copy(h2.attrs, h.attrs)
	copy(h2.attrs[len(h.attrs):], attrs)
	return h2
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	// Create a new handler; do not copy the mutex.
	h2 := &ConsoleHandler{
		opts:  h.opts,
		w:     h.w,
		attrs: h.attrs, // Attributes are carried over.
	}
	if h.group == "" {
		h2.group = name
	} else {
		h2.group = h.group + "." + name
	}
	return h2
}
