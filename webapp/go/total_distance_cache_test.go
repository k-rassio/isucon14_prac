package main

import (
	"testing"
	"time"
)

func TestTotalDistanceCacheRoundTrip(t *testing.T) {
	resetTotalDistanceCache()

	expectedAt := time.Unix(1700000000, 0).UTC()
	setTotalDistanceCacheValue("chair-1", 1234, expectedAt)

	distance, updatedAt, ok := getTotalDistanceCacheValue("chair-1")
	if !ok {
		t.Fatalf("expected cache entry to exist")
	}
	if distance != 1234 {
		t.Fatalf("unexpected distance: got %d want %d", distance, 1234)
	}
	if !updatedAt.Equal(expectedAt) {
		t.Fatalf("unexpected updatedAt: got %v want %v", updatedAt, expectedAt)
	}
}

func TestTotalDistanceCacheMissingEntry(t *testing.T) {
	resetTotalDistanceCache()

	_, _, ok := getTotalDistanceCacheValue("missing-chair")
	if ok {
		t.Fatalf("expected missing cache entry")
	}
}
