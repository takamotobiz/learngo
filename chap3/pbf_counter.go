package osmpbfparser

import (
	"github.com/thomersch/gosmparse"
	"os"
	"sync"
)

// PBFCounter count the number of each type data element and return the number.
type PBFCounter struct {
	PBFFile       string
	NodeCount     int
	WayCount      int
	RelationCount int
	nodeMux       sync.Mutex
	wayMux        sync.Mutex
	relationMux   sync.Mutex
}

// Run exec func for PBFCounter.
func (p *PBFCounter) Run() (int, int, int, error) {
	reader, err := os.Open(p.PBFFile)
	if err != nil {
		return 0, 0, 0, err
	}
	defer reader.Close()

	decoder := gosmparse.NewDecoder(reader)
	if err := decoder.Parse(p); err != nil {
		return 0, 0, 0, err
	}
	return p.NodeCount, p.WayCount, p.RelationCount, nil
}

// ReadNode for gosmparse OSMReader interface.
func (p *PBFCounter) ReadNode(node gosmparse.Node) {
	p.nodeMux.Lock()
	p.NodeCount++
	p.nodeMux.Unlock()
}

// ReadWay for gosmparse OSMReader interface.
func (p *PBFCounter) ReadWay(way gosmparse.Way) {
	p.wayMux.Lock()
	p.WayCount++
	p.wayMux.Unlock()
}

// ReadRelation for gosmparse OSMReader interface.
func (p *PBFCounter) ReadRelation(relation gosmparse.Relation) {
	p.relationMux.Lock()
	p.RelationCount++
	p.relationMux.Unlock()
}
