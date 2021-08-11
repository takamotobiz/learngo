// +build integration

package osmpbfparser

import (
	"github.com/jneo8/osmpbfparser-go/bitmask"
	"github.com/vbauerster/mpb/v5"
	"sync"
	"testing"
)

func TestPBFIndexer(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)
	pb := mpb.New(mpb.WithWaitGroup(&wg))
	indexer := newPBFIndexer(
		"./static/test.pbf",
		bitmask.NewPBFMasks(),
		pb.AddBar(int64(1115337)),
		pb.AddBar(int64(50832)),
		pb.AddBar(int64(243)),
	)
	if err := indexer.Run(); err != nil {
		t.Error(err)
	}
	wg.Done()
	wg.Done()
	wg.Done()
	pb.Wait()
}
