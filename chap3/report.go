package osmpbfparser

import (
	"fmt"
	"time"
)

// Report recor the information duration the runtime.
type Report struct {
	PBFFile           string
	Fizesize          int64
	SpendTime         time.Duration
	ProcessedRelation int
	ProcessedWay      int
	ProcessedNode     int
	FatalRelation     int
	FatalWay          int
}

// GetReport return report string.
func (r *Report) GetReport() string {
	return fmt.Sprintf(
		`
			PBF: %s
			FileSize: %d MB
			Timeit: %2f Secs
			ProcessRelation: %d,
			ProcessWay: %d,
			ProcessNode: %d,
			FatalRelation: %d,
			FatalWay: %d,
		`,
		r.PBFFile,
		r.Fizesize/(1024*1024),
		r.SpendTime.Seconds(),
		r.ProcessedRelation,
		r.ProcessedWay,
		r.ProcessedNode,
		r.FatalRelation,
		r.FatalWay,
	)
}
