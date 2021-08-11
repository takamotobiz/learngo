// +build geojson_integration

package osmpbfparser

import (
	"github.com/paulmach/go.geojson"
	"testing"
)

func TestElementToGeojson(t *testing.T) {
	parser := New(
		Args{
			PBFFile:     "./static/test.pbf",
			LevelDBPath: "/tmp/osmpbfparser",
			BatchSize:   10000,
		},
	)

	for e := range parser.Iterator() {
		rawJSON, err := e.ToGeoJSON()
		if err != nil {
			t.Error(err)
		}
		if _, err := geojson.UnmarshalFeatureCollection(rawJSON); err != nil {
			t.Errorf("%s Element %d got %s ", err, e.GetID(), string(rawJSON))
			break
		}
	}
}
