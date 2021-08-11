// +build integration

package osmpbfparser

import (
	"github.com/jneo8/osmpbfparser-go/bitmask"
	"testing"
)

func TestPBFRelationMemberIndexer(t *testing.T) {

	masks := bitmask.NewPBFMasks()
	pb, wg := newProgress(6)

	indexer := newPBFIndexer(
		"./static/test.pbf",
		masks,
		addBar(pb, "IndexerNode", 1115337),
		addBar(pb, "IndexerWay", 50832),
		addBar(pb, "IndexerRelation", 243),
	)
	if err := indexer.Run(); err != nil {
		t.Error(err)
	}
	wg.Done()
	wg.Done()
	wg.Done()

	rmIndexer := newPBFRelationMemberIndexer(
		"./static/test.pbf",
		masks,
		addBar(pb, "IndexerNode", 1115337),
		addBar(pb, "IndexerWay", 50832),
		addBar(pb, "IndexerRelation", 243),
	)
	if err := rmIndexer.Run(); err != nil {
		t.Error(err)
	}
	wg.Done()
	wg.Done()
	wg.Done()
	pb.Wait()
}
