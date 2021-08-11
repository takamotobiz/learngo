package osmpbfparser

import (
	log "github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
)

// PBFParser is a item parse pbf file, iterator return.
type PBFParser interface {
	Iterator() <-chan Element
	SetLogger(*log.Logger)
	Err() error
	Close() error
}

type pbfDataParser interface {
	gosmparse.OSMReader
	Run() error
}

type pbfDataCounter interface {
	gosmparse.OSMReader
	Run() (nodeCount int, wayCount int, relationCount int, err error)
}
