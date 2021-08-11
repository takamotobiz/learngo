# OSM PBF Parser

- Parse pbf file to osm element(Node, Way and Relation)

- Iterator pattern. Easy to get.

- Progress bar for parsing process.

- Control memory usage by using levelDB.

- Convert Element to GeoJSON
    - Support type: Node, Way, Relation(polygon, multipolygon)

## Install

```
go get -u https://github.com/jneo8/osmpbfparser-go
```

## Quick start

```bash
# This will run the example code in ./cmd/example/main.go
make run-example
```
