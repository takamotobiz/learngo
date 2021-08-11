package osmpbfparser

import (
// "github.com/jneo8/osmpbfparser-go/bitmask"
)

// Args is inpute arguments struct.
type Args struct {
	PBFFile     string // pbf file path.
	LevelDBPath string // levelDB init path.
	BatchSize   int    // levelDB batch size, it depends on how many memory you have.
}
