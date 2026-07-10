package main

import (
	"strings"
	"testing"
	"time"
)

func TestGuardEnforcesGapBetweenSends(t *testing.T) {
	g := &sendGuard{}
	if err := g.admit(); err != nil {
		t.Fatalf("first admit should pass: %v", err)
	}
	start := time.Now()
	if err := g.admit(); err != nil {
		t.Fatalf("second admit should pass: %v", err)
	}
	if gap := time.Since(start); gap < guardMinGap {
		t.Errorf("second send admitted after %v, want >= %v", gap, guardMinGap)
	}
}

func TestGuardHourlyCap(t *testing.T) {
	g := &sendGuard{}
	now := time.Now()
	// Simulate a full window of recent sends without sleeping.
	for i := 0; i < guardMaxPerHr; i++ {
		g.sendTimes = append(g.sendTimes, now.Add(-time.Duration(i)*time.Minute))
	}
	g.lastSend = now
	err := g.admit()
	if err == nil {
		t.Fatal("admit should fail once the hourly cap is reached")
	}
	if !strings.Contains(err.Error(), "hourly cap") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGuardWindowExpiry(t *testing.T) {
	g := &sendGuard{}
	// A window full of OLD sends (>1h) must not count against the cap.
	old := time.Now().Add(-2 * time.Hour)
	for i := 0; i < guardMaxPerHr; i++ {
		g.sendTimes = append(g.sendTimes, old)
	}
	if err := g.admit(); err != nil {
		t.Fatalf("stale window should have expired, got: %v", err)
	}
	if len(g.sendTimes) != 1 {
		t.Errorf("expected pruned window with 1 entry, got %d", len(g.sendTimes))
	}
}
