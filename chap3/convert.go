package osmpbfparser

import (
	"bytes"
	"encoding/binary"
	"github.com/thomersch/gosmparse"
	"math"
	"strconv"
)

// nodeToBytes convert node struct to bytes.
func nodeToBytes(n gosmparse.Node) (string, []byte) {
	var buf bytes.Buffer

	var latBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(latBytes, math.Float64bits(n.Lat))
	buf.Write(latBytes)

	var lngBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(lngBytes, math.Float64bits(n.Lon))
	buf.Write(lngBytes)

	return strconv.FormatInt(n.ID, 10), buf.Bytes()
}

// bytesToNode convert bytes to node struct.
func bytesToNode(b []byte) gosmparse.Node {
	node := gosmparse.Node{}

	var latBytes = append([]byte{}, b[0:8]...)
	var lat = math.Float64frombits(binary.BigEndian.Uint64(latBytes))
	node.Lat = lat

	var lngBytes = append([]byte{}, b[8:16]...)
	var lng = math.Float64frombits(binary.BigEndian.Uint64(lngBytes))
	node.Lon = lng
	return node
}
