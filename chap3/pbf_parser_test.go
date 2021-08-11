// +build integration

package osmpbfparser

import (
	"testing"
)

// TestPBFParser ...
func TestPBFParser(t *testing.T) {
	parser := New(
		Args{
			PBFFile:     "./static/test.pbf",
			LevelDBPath: "/tmp/osmpbfparser",
			BatchSize:   10000,
		},
	)

	for range parser.Iterator() {
		// ignore return item
	}
	if err := parser.Err(); err != nil {
		t.Error(err)
	}
}
