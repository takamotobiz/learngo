package osmpbfparser

import (
	"github.com/jneo8/logger-go"
	"github.com/jneo8/osmpbfparser-go/bitmask"
	"github.com/vbauerster/mpb/v5"
)

// New return new PBFParser.
func New(
	args Args,
) PBFParser {
	return &pbfParser{
		Args:   args,
		Logger: logger.NewLogger(),
		Report: Report{PBFFile: args.PBFFile},
	}
}

func newPBFIndexer(
	pbfFile string,
	pbfMasks *bitmask.PBFMasks,
	nodeBar *mpb.Bar,
	wayBar *mpb.Bar,
	relationBar *mpb.Bar,
) pbfDataParser {
	return &PBFIndexer{
		PBFFile:     pbfFile,
		PBFMasks:    pbfMasks,
		NodeBar:     nodeBar,
		WayBar:      wayBar,
		RelationBar: relationBar,
	}
}

func newPBFRelationMemberIndexer(
	pbfFile string,
	pbfMasks *bitmask.PBFMasks,
	nodeBar *mpb.Bar,
	wayBar *mpb.Bar,
	relationBar *mpb.Bar,
) pbfDataParser {
	return &PBFRelationMemberIndexer{
		PBFFile:     pbfFile,
		PBFMasks:    pbfMasks,
		NodeBar:     nodeBar,
		WayBar:      wayBar,
		RelationBar: relationBar,
	}
}

func newPBFCounter(
	pbfFile string,
) pbfDataCounter {
	return &PBFCounter{
		PBFFile: pbfFile,
	}
}
