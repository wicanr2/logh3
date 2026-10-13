package main

import (
	"sync/atomic"
	"testing"
	"time"
)

type startupMusic struct {
	stops atomic.Int32
	done  chan struct{}
}

func (m *startupMusic) Play(string) error { return nil }
func (m *startupMusic) Playing() string   { return "" }
func (m *startupMusic) Stop() {
	if m.stops.Add(1) == 1 {
		close(m.done)
	}
}

func TestStartupCloseDuringLoadCleansCompletedPlayer(t *testing.T) {
	entered, finish := make(chan struct{}), make(chan struct{})
	player := &startupMusic{done: make(chan struct{})}
	s := &startup{load: func(report func(int, string)) (*app, error) {
		report(1, "載入資料")
		close(entered)
		<-finish
		return &app{music: player}, nil
	}}
	s.begin()
	<-entered
	s.close()
	close(finish)
	select {
	case <-player.done:
	case <-time.After(5 * time.Second):
		t.Fatal("載入視窗關閉後未收播放器")
	}
	s.close()
	if player.stops.Load() != 1 {
		t.Fatal("播放器重複收尾")
	}
}
