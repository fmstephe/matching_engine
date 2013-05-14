package main

import (
	"testing"
)

// This test doesn't test anything other than that the perf tool runs and does not deadlock
func TestPerf(t *testing.T) {
	doPerf(false)
}
