package osmpbfparser

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/thomersch/gosmparse"
)

// BytesToElement convert bytes to Element struct.
func BytesToElement(b []byte) (Element, error) {
	decoder := gob.NewDecoder(bytes.NewReader(b))
	var element Element
	err := decoder.Decode(&element)
	return element, err
}

// Element is an osm data element group set.
type Element struct {
	Type     int // 0=Node, 1=Way. 2=Relation
	Node     gosmparse.Node
	Way      gosmparse.Way
	Relation gosmparse.Relation
	Elements []Element
	Role     int // 0=outer, 1=inner
}

// ToBytes convert element struct to bytes.
func (e *Element) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(e); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// GetID return ID.
func (e *Element) GetID() int64 {
	var id int64
	switch e.Type {
	case 0:
		id = e.Node.ID
	case 1:
		id = e.Way.ID
	case 2:
		id = e.Relation.ID
	}
	return id
}

// IsArea return if element is area or not.
// https://wiki.openstreetmap.org/wiki/Key:area
func (e *Element) IsArea() bool {
	var isPolygon bool
	if val, ok := e.Way.Tags["area"]; ok && val == "yes" {
		// This list is probably incomplete - please add other cases
		if _, ok := e.Way.Tags["highway"]; ok {
			// highway=*, see 'Highway areas' below for more details
		} else if _, isBarrier := e.Way.Tags["barrier"]; isBarrier {
			// barrier=*, for thicker hedges or walls or detailed mapping defined using an area add area=yes
		} else {
			isPolygon = true
		}
	}
	return isPolygon
}

// GetTags ...
func (e *Element) GetTags() map[string]string {
	switch e.Type {
	case 0:
		return e.Node.Tags
	case 1:
		return e.Way.Tags
	case 2:
		return e.Relation.Tags
	}
	return make(map[string]string)
}

// GetName from tags.
func (e *Element) GetName() (string, error) {
	tags := e.GetTags()
	if v, ok := tags["name"]; ok {
		return v, nil
	}
	return "", fmt.Errorf("No name tag")
}
