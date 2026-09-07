// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	libtime "github.com/bborbe/time"
	"github.com/gorilla/mux"
)

// NewSetLogLevelHandler returns a handler for the /setloglevel/{level} route.
// The numeric level segment is validated (0-10) and mapped onto the given
// slog LevelVar (0 -> Info, >=1 -> Debug); invalid input yields a 400
// response and leaves the level unchanged. Unless the level was re-set in
// the meantime, it auto-resets to defaultLevel after autoResetDuration.
func NewSetLogLevelHandler(
	levelVar *slog.LevelVar,
	defaultLevel slog.Level,
	autoResetDuration time.Duration,
) http.Handler {
	return &logLevelSetter{
		levelVar:          levelVar,
		defaultLevel:      defaultLevel,
		autoResetDuration: autoResetDuration,
	}
}

type logLevelSetter struct {
	levelVar          *slog.LevelVar
	defaultLevel      slog.Level
	autoResetDuration time.Duration

	mux         sync.Mutex
	lastSetTime time.Time
}

func (l *logLevelSetter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	n, err := strconv.Atoi(mux.Vars(req)["level"])
	if err != nil || n < 0 || n > 10 {
		http.Error(resp, "invalid loglevel", http.StatusBadRequest)
		return
	}
	level := slog.LevelInfo
	if n >= 1 {
		level = slog.LevelDebug
	}
	l.setLevel(level)
	slog.Info("set loglevel", "level", n, "reset_in", l.autoResetDuration)
	fmt.Fprintf(resp, "set loglevel to %d completed\n", n)
}

func (l *logLevelSetter) setLevel(level slog.Level) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.lastSetTime = libtime.Now()
	l.levelVar.Set(level)
	go l.resetLater()
}

func (l *logLevelSetter) resetLater() {
	time.Sleep(l.autoResetDuration)
	l.mux.Lock()
	defer l.mux.Unlock()
	if libtime.Now().Sub(l.lastSetTime) < l.autoResetDuration {
		slog.Debug("skip loglevel reset, set more recently")
		return // re-set in the meantime; keep the newer level
	}
	slog.Info("log level reset to default", "level", l.defaultLevel)
	l.levelVar.Set(l.defaultLevel)
}
