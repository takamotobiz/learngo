package osmpbfparser

import (
	"reflect"
	"strconv"

	geojson "github.com/paulmach/go.geojson"
)

// ToGeoJSON convery element to JSON bytes.
func (e *Element) ToGeoJSON() ([]byte, error) {
	switch e.Type {
	}
	// fc := geojson.NewFeatureCollection()
	f := e.ToGeoJSONFeature()
	// fc.AddFeature(f)
	// rawJSON, err := fc.MarshalJSON()
	rawJSON, err := f.MarshalJSON()
	return rawJSON, err
}

// ToGeoJSONFeature ...
func (e *Element) ToGeoJSONFeature() *geojson.Feature {
	var f *geojson.Feature
	switch e.Type {
	case 0:
		f = e.NodeToFeature()
	case 1:
		f = e.WayToFeature()
	case 2:
		f = e.RelationToFeature()
	}
	return f
}

// WayToFeature convert way element to geojson feature.
func (e *Element) WayToFeature() *geojson.Feature {
	latLngs := [][]float64{}
	for _, member := range e.Elements {
		latLngs = append(latLngs, []float64{member.Node.Lon, member.Node.Lat})
	}

	var f *geojson.Feature

	switch e.IsArea() {
	case true:
		f = geojson.NewPolygonFeature([][][]float64{latLngs})
	default:
		f = geojson.NewLineStringFeature(latLngs)
	}

	wayID := "way" + "/" + strconv.FormatInt(e.Way.ID, 10)
	f.ID = wayID
	f.SetProperty("osmid", wayID)
	f.SetProperty("osmType", "way")

	for k, v := range e.Way.Tags {
		f.SetProperty(k, v)
	}
	return f
}

// NodeToFeature convert node element to geojson feature.
func (e *Element) NodeToFeature() *geojson.Feature {
	f := geojson.NewPointFeature([]float64{e.Node.Lon, e.Node.Lat})

	nodeID := "node/" + strconv.FormatInt(e.Node.ID, 10)
	f.ID = nodeID
	f.SetProperty("osmid", nodeID)
	f.SetProperty("osmType", "node")

	for k, v := range e.Node.Tags {
		f.SetProperty(k, v)
	}
	return f
}

// IsMultiPolygon check if element is multi-polygon or not.
func (e *Element) IsMultiPolygon() bool {
	if e.Type == 2 {
		if v, ok := e.Relation.Tags["type"]; ok {
			if v == "multipolygon" {
				return true
			}
		}
	}
	return false
}

// RelationToFeature convert relation element to geojson feature.
func (e *Element) RelationToFeature() *geojson.Feature {
	var f *geojson.Feature
	if e.IsMultiPolygon() {
		f = e.GetRelationAsMultipolygon()
	} else {
		f = e.GetRelationAsGeometries()
	}
	id := "relation/" + strconv.FormatInt(e.Relation.ID, 10)
	f.ID = id
	f.SetProperty("osmid", id)
	f.SetProperty("osmType", "relation")
	for k, v := range e.Relation.Tags {
		f.SetProperty(k, v)
	}
	return f
}

// GetRelationAsMultipolygon return relation as geosjon multipolygon.
func (e *Element) GetRelationAsMultipolygon() *geojson.Feature {

	multiPolygon := [][][][]float64{}
	points := [][]float64{}
	emtPoints := [][]float64{}
	polygon := [][][]float64{}
	innerPolygon := [][][]float64{}
	for _, emt := range e.Elements {
		emtFeature := emt.ToGeoJSONFeature()
		switch emtFeature.Geometry.Type {
		case geojson.GeometryPoint:
			emtPoints = append(emtPoints, emtFeature.Geometry.Point)
		case geojson.GeometryMultiPoint:
			emtPoints = append(emtPoints, emtFeature.Geometry.MultiPoint...)
		case geojson.GeometryLineString:
			emtPoints = append(emtPoints, emtFeature.Geometry.LineString...)
		case geojson.GeometryMultiLineString:
			for _, lineString := range emtFeature.Geometry.MultiLineString {
				emtPoints = append(emtPoints, lineString...)
			}
		case geojson.GeometryPolygon:
			multiPolygon = append(multiPolygon, emtFeature.Geometry.Polygon)
		case geojson.GeometryMultiPolygon:
			multiPolygon = append(multiPolygon, emtFeature.Geometry.MultiPolygon...)
		}

		points = multiPolygonPointsAppendPoints(points, emtPoints)
		multiPolygon, polygon, points, innerPolygon = multiAreaMultiPolygonAppend(multiPolygon, polygon, points, emt.Role, innerPolygon)
	}
	// Final flush
	if len(polygon) > 0 {
		if len(innerPolygon) > 0 {
			polygon = append(polygon, innerPolygon...)
		}
		multiPolygon = append(multiPolygon, polygon)
	}
	return geojson.NewMultiPolygonFeature(multiPolygon...)
}

// Area multi-polygon append role.
// Check is area and append by different role(inner, outer) cases.
func multiAreaMultiPolygonAppend(multiPolygon [][][][]float64, polygon [][][]float64, points [][]float64, role int, innerPolygon [][][]float64) ([][][][]float64, [][][]float64, [][]float64, [][][]float64) {
	// Is area?
	// Is points length > 1 && first point equal to last point.
	if len(points) > 1 && reflect.DeepEqual(points[0], points[len(points)-1]) {
		switch role {
		case 0: // outer
			if len(polygon) > 0 {
				multiPolygon = append(multiPolygon, polygon)
				polygon = [][][]float64{}
			}
			polygon = append(polygon, points)
			if len(innerPolygon) > 0 {
				polygon = append(polygon, innerPolygon...)
				innerPolygon = [][][]float64{}
			}
		case 1: // inner
			if len(polygon) == 0 {
				innerPolygon = append(innerPolygon, points)
			} else {
				polygon = append(polygon, points)
			}
		}
		points = [][]float64{}
	}
	return multiPolygon, polygon, points, innerPolygon

}

// Checkint the graft point, different case will have different way to append.
// Case (AEnd, BStart), (AEnd, BEnd), (AStart, BEnd), (AStart, BStart)
func multiPolygonPointsAppendPoints(pointsA [][]float64, pointsB [][]float64) [][]float64 {
	if len(pointsB) == 0 {
		return pointsA
	}
	if len(pointsA) == 0 {
		return pointsB
	}
	var bStart, bEnd []float64
	if len(pointsB) > 0 {
		bStart = pointsB[0]
		bEnd = pointsB[len(pointsB)-1]
	}

	aStart := pointsA[0]
	aEnd := pointsA[len(pointsA)-1]
	if reflect.DeepEqual(aEnd, bStart) {
		pointsA = append(pointsA, pointsB...)
	} else if reflect.DeepEqual(aStart, bStart) {
		newPoints := [][]float64{}
		for i := len(pointsB) - 1; i >= 0; i-- {
			newPoints = append(newPoints, pointsB[i])
		}
		pointsA = append(newPoints, pointsA...)
	} else if reflect.DeepEqual(aEnd, bEnd) {
		for i := len(pointsB) - 1; i >= 0; i-- {
			pointsA = append(pointsA, pointsB[i])
		}
	} else {
		pointsA = append(pointsB, pointsA...)
	}
	return pointsA
}

// GetRelationAsGeometries return relation as geojson geometries.
// Sometimes relation will missing type tag, so defualt we use collect feature.
func (e *Element) GetRelationAsGeometries() *geojson.Feature {
	geometries := []*geojson.Geometry{}
	for _, emt := range e.Elements {
		switch emt.Type {
		case 0:
			geometries = append(geometries, emt.NodeToFeature().Geometry)
		case 1:
			geometries = append(geometries, emt.WayToFeature().Geometry)
		case 2:
			geometries = append(geometries, emt.RelationToFeature().Geometry)
		}
	}
	return geojson.NewCollectionFeature(geometries...)
}
