package osmpbfparser

import (
	"github.com/jneo8/osmpbfparser-go/bitmask"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/thomersch/gosmparse"
	"github.com/vbauerster/mpb/v5"
	"os"
	"strconv"
	"sync"
	"time"
)

type pbfParser struct {
	PBFMasks *bitmask.PBFMasks

	// leveldb
	DB    *leveldb.DB
	Batch *leveldb.Batch
	Args  Args

	// Log
	Logger    *log.Logger
	elementCh chan Element

	Error error

	// Report
	Report Report

	// output
	OutputCh chan Element
}

func (p *pbfParser) Err() error {
	return p.Error
}

func (p *pbfParser) Close() error {
	return os.RemoveAll(p.Args.LevelDBPath)
}

// Run ...
func (p *pbfParser) Iterator() <-chan Element {
	p.OutputCh = make(chan Element)

	go func() {
		defer close(p.OutputCh)
		st := time.Now()

		// Get file size
		reader, err := os.Open(p.Args.PBFFile)
		if err != nil {
			p.Error = err
			return
		}
		defer reader.Close()
		fInfo, err := reader.Stat()
		if err != nil {
			p.Error = err
			return
		}
		p.Report.Fizesize = fInfo.Size()
		reader.Close()

		db, err := leveldb.OpenFile(
			p.Args.LevelDBPath,
			&opt.Options{DisableBlockCache: true},
		)
		if err != nil {
			p.Logger.Error(err)
			p.Error = err
			return
		}
		defer db.Close()
		p.DB = db
		p.Logger.Infof("Init leveldb at %s", p.Args.LevelDBPath)

		// bitmask
		p.PBFMasks = bitmask.NewPBFMasks()

		// counter
		p.Logger.Info("Start Count")
		counter := newPBFCounter(
			p.Args.PBFFile,
		)
		nodeC, wayC, relationC, err := counter.Run()
		if err != nil {
			p.Error = err
			return
		}
		p.Logger.Info(nodeC, wayC, relationC)

		// Init bar
		pb, wg := newProgress(12)

		// Index
		indexer := newPBFIndexer(
			p.Args.PBFFile,
			p.PBFMasks,
			addBar(pb, "IndexerNode", nodeC),
			addBar(pb, "IndexerWay", wayC),
			addBar(pb, "IndexerRelation", relationC),
		)
		if err := indexer.Run(); err != nil {
			p.Error = err
			return
		}
		wg.Done()
		wg.Done()
		wg.Done()

		// Relation member indexer
		relationMemberIndexer := newPBFRelationMemberIndexer(
			p.Args.PBFFile,
			p.PBFMasks,
			addBar(pb, "RMIndexerNode", nodeC),
			addBar(pb, "RMIndexerWay", wayC),
			addBar(pb, "RMIndexerRelation", relationC),
		)
		if err := relationMemberIndexer.Run(); err != nil {
			p.Error = err
			return
		}
		wg.Done()
		wg.Done()
		wg.Done()

		// First round
		// Insert element to leveldb
		if err := p.runInserter(
			wg,
			addBar(pb, "insertNode", nodeC),
			addBar(pb, "insertWay", wayC),
			addBar(pb, "insertRelation", relationC),
		); err != nil {
			p.Error = err
			return
		}

		// Final round.
		// Output element
		if err := p.runOutputer(
			wg,
			addBar(pb, "outputNode", nodeC),
			addBar(pb, "outputWay", wayC),
			addBar(pb, "outputRelation", relationC),
		); err != nil {
			p.Error = err
			return
		}

		// Report
		pb.Wait()

		p.Report.SpendTime = time.Since(st)
		p.Logger.Infof(p.Report.GetReport())
	}()
	return p.OutputCh
}

// ReadNode for gosmparse OSMReader interface.
func (p *pbfParser) ReadNode(node gosmparse.Node) {
	p.elementCh <- Element{
		Type: 0,
		Node: node,
	}
}

// ReadWay for gosmparse OSMReader interface.
func (p *pbfParser) ReadWay(way gosmparse.Way) {
	p.elementCh <- Element{
		Type: 1,
		Way:  way,
	}
}

// ReadRelation for gosmparse OSMReader interface.
func (p *pbfParser) ReadRelation(relation gosmparse.Relation) {
	p.elementCh <- Element{
		Type:     2,
		Relation: relation,
	}
}

// SetLogger ...
func (p *pbfParser) SetLogger(logger *log.Logger) {
	p.Logger = logger
}

// InsertDB, Put way refs, relation member into db.
func (p *pbfParser) runInserter(
	wg *sync.WaitGroup,
	nodeBar *mpb.Bar,
	wayBar *mpb.Bar,
	relationBar *mpb.Bar,
) error {
	defer wg.Done()
	defer wg.Done()
	defer wg.Done()

	// reader .
	reader, err := os.Open(p.Args.PBFFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	p.Batch = leveldb.MakeBatch(p.Args.BatchSize)
	defer p.Batch.Reset()

	p.elementCh = make(chan Element)

	insertDBWg := sync.WaitGroup{}
	insertDBWg.Add(1)
	go func() {
		defer insertDBWg.Done()
		idx := 0
		for emt := range p.elementCh {
			idx++
			// Check batch size every batchsize.
			if idx%p.Args.BatchSize == 0 {
				if err := p.checkBatch(); err != nil {
					p.Error = err
				}
			}
			switch emt.Type {
			case 0:
				nodeBar.Increment()
				if p.PBFMasks.WayRefs.Has(emt.Node.ID) || p.PBFMasks.RelNodes.Has(emt.Node.ID) {
					id, b := nodeToBytes(emt.Node)
					p.Batch.Put(
						[]byte(id),
						b,
					)
				}
			case 1:
				wayBar.Increment()
				if p.PBFMasks.RelWays.Has(emt.Way.ID) {
					emtBytes, err := emt.ToBytes()
					if err != nil {
						p.Report.FatalWay++
						continue
					}
					p.Batch.Put(
						[]byte("W"+strconv.FormatInt(emt.Way.ID, 10)),
						emtBytes,
					)
				}
			case 2:
				relationBar.Increment()
				if p.PBFMasks.RelRelation.Has(emt.Relation.ID) {
					emtBytes, err := emt.ToBytes()
					if err != nil {
						p.Report.FatalRelation++
						continue
					}
					p.Batch.Put(
						[]byte("R"+strconv.FormatInt(emt.Relation.ID, 10)),
						emtBytes,
					)
				}
			}
		}
	}()

	firstRoundDecoder := gosmparse.NewDecoder(reader)
	if err := firstRoundDecoder.Parse(p); err != nil {
		return err
	}
	close(p.elementCh)
	insertDBWg.Wait()

	// Flush batch.
	if err := p.flushBatch(true); err != nil {
		return err
	}
	return nil
}

func (p *pbfParser) runOutputer(
	wg *sync.WaitGroup,
	nodeBar *mpb.Bar,
	wayBar *mpb.Bar,
	relationBar *mpb.Bar,
) error {
	defer wg.Done()
	defer wg.Done()
	defer wg.Done()

	reader, err := os.Open(p.Args.PBFFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Init element chan
	p.elementCh = make(chan Element)

	finalRWg := sync.WaitGroup{}
	finalRWg.Add(1)
	go func() {
		defer finalRWg.Done()
		for emt := range p.elementCh {
			switch emt.Type {
			case 0:
				nodeBar.Increment()
				if p.PBFMasks.Nodes.Has(emt.Node.ID) {
					p.Report.ProcessedNode++
					p.OutputCh <- emt
				}
			case 1:
				wayBar.Increment()
				if p.PBFMasks.Ways.Has(emt.Way.ID) {
					emts, err := p.dbLookupWayEmts(&emt.Way)
					if err != nil {
						p.Logger.Warning(err)
						p.Report.FatalWay++
						continue
					}
					emt.Elements = emts
					p.Report.ProcessedWay++
					p.OutputCh <- emt
				}

			case 2:
				relationBar.Increment()
				if p.PBFMasks.Relations.Has(emt.Relation.ID) {
					emts, err := p.dbLookupRelationEmts(&emt.Relation, []int64{})
					if err != nil {
						// p.Logger.Warning(err)
						p.Report.FatalRelation++
						continue
					}
					emt.Elements = emts
					p.Report.ProcessedRelation++
					p.OutputCh <- emt
				}
			}
		}
	}()
	decoder := gosmparse.NewDecoder(reader)
	if err := decoder.Parse(p); err != nil {
		return err
	}
	close(p.elementCh)
	finalRWg.Wait()
	return nil
}

func (p *pbfParser) checkBatch() error {
	if p.Batch.Len() >= p.Args.BatchSize {
		if err := p.flushBatch(true); err != nil {
			return err
		}
	}
	return nil
}

func (p *pbfParser) flushBatch(sync bool) error {
	writeOpts := opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         sync,
	}
	if err := p.DB.Write(p.Batch, &writeOpts); err != nil {
		return err
	}
	p.Batch.Reset()
	return nil
}

func (p *pbfParser) dbLookupNodeElementByID(id int64) (Element, error) {
	b, err := p.DB.Get(
		[]byte(strconv.FormatInt(id, 10)),
		nil,
	)
	if err != nil {
		return Element{}, err
	}
	node := bytesToNode(b)
	emt := Element{
		Type: 0,
		Node: node,
	}
	return emt, nil
}

func (p *pbfParser) dbLookupWayElementByID(id int64) (Element, error) {
	b, err := p.DB.Get(
		[]byte("W"+strconv.FormatInt(id, 10)),
		nil,
	)
	if err != nil {
		return Element{}, err
	}
	emt, err := BytesToElement(b)
	if err != nil {
		return emt, err
	}
	return emt, nil
}

func (p *pbfParser) dbLookupRelationElementByID(id int64) (Element, error) {
	b, err := p.DB.Get(
		[]byte("R"+strconv.FormatInt(id, 10)),
		nil,
	)
	if err != nil {
		return Element{}, err
	}
	emt, err := BytesToElement(b)
	if err != nil {
		return emt, err
	}
	return emt, nil
}

func (p *pbfParser) dbLookupWayEmts(way *gosmparse.Way) ([]Element, error) {
	var emts []Element
	for _, nodeID := range way.NodeIDs {
		emt, err := p.dbLookupNodeElementByID(nodeID)
		if err != nil {
			return emts, err
		}
		emts = append(emts, emt)
	}
	return emts, nil
}

func (p *pbfParser) dbLookupRelationEmts(relation *gosmparse.Relation, processedList []int64) ([]Element, error) {
	var emts []Element
	processedList = append(processedList, relation.ID)
	for _, member := range relation.Members {
		var element Element
		switch member.Type {
		case 0:
			emt, err := p.dbLookupNodeElementByID(member.ID)
			if err != nil {
				return emts, err
			}
			element = emt
		case 1:
			emt, err := p.dbLookupWayElementByID(member.ID)
			if err != nil {
				return emts, err
			}

			memberEmts, err := p.dbLookupWayEmts(&emt.Way)
			if err != nil {
				return emts, err
			}
			emt.Elements = memberEmts
			element = emt
		case 2:
			// Passing if already processed.
			var processed bool
			for _, processedID := range processedList {
				if member.ID == processedID {
					processed = true
				}
			}
			if processed {
				continue
			}
			emt, err := p.dbLookupRelationElementByID(member.ID)
			if err != nil {
				return emts, err
			}
			memberEmts, err := p.dbLookupRelationEmts(&emt.Relation, processedList)
			if err != nil {
				return emts, err
			}
			emt.Elements = memberEmts
			element = emt
		}
		switch member.Role {
		case "inner":
			element.Role = 1
		default:
			// default is outer
			element.Role = 0
		}
		emts = append(emts, element)
	}
	return emts, nil
}
