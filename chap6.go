package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

// define file
// const pbfname string = "/data/osm.pbf/shikoku-low.osm.pbf"
const pbfname string = "/data/osm.pbf/japan-low.osm.pbf"

// const pbfname string = "/data/osm.pbf/planet-low.osm.pbf"
// const pbfname string = "/data/osm.pbf/hokkaido-low.osm.pbf"
const outfname string = "./output.json"

// define Tag
const tagname string = "amenity"
const tagval string = "school"

func main() {

	// get start time
	start := time.Now()

	// ====================================
	// 1-1.open osm.pbf file and create scanner
	// ====================================
	home, _ := os.UserHomeDir()
	f, err := os.Open(home + pbfname)
	if err != nil {
		fmt.Printf("could not open file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	// get number of cpu and create scanner
	cpu := runtime.NumCPU()
	scanner := osmpbf.New(context.Background(), f, cpu)

	// ====================================
	// 1-1.create Node list include Relations(mrway)
	// ====================================
	scanner.SkipNodes = true
	scanner.SkipWays = true
	mrway := map[int]string{}
	// for debug
	mdebug := map[string]int{}
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Relation:
			if e.Tags.Find(tagname) == tagval {
				for _, v := range e.Members {
					k := e.Tags.Find("type") + "/" + string(v.Type) + "/" + v.Role + "/"
					mrway[int(v.Ref)] = k
					// for debug
					mdebug[k] += 1
				}
			}
		}
	}
	scanner.Close()
	fmt.Println("mdebug[", mdebug, "]")

	// ====================================
	// 1-2.add coordinate and set open/close to mrway
	// ====================================
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch re := scanner.Object().(type) {
		case *osm.Way:
			// set way coordinates
			if _, flg := mrway[int(re.ID)]; flg {
				//fmt.Println("Way:", int(e.ID))
				var flon, flat, llon, llat float64
				for i, v := range re.Nodes {
					if i == 0 {
						mrway[int(re.ID)] += fmt.Sprintf("[%.7f,%.7f]", v.Lon, v.Lat)
						flon, flat = v.Lon, v.Lat
					} else {
						mrway[int(re.ID)] += fmt.Sprintf(",[%.7f,%.7f]", v.Lon, v.Lat)
						llon, llat = v.Lon, v.Lat
					}
				}
				// first and last coordinate
				mrway[int(re.ID)] += fmt.Sprintf("[%.7f,%.7f]", flon, flat)
				mrway[int(re.ID)] += fmt.Sprintf("[%.7f,%.7f]", llon, llat)
				//
				if llon == flon && llat == flat {
					mrway[int(re.ID)] += "/close"
				} else {
					mrway[int(re.ID)] += "/open"
				}
			}
		}
	}
	scanner.Close()

	// ====================================
	// 2.create GeoJSON file
	// ====================================
	file, err := os.Create(outfname)
	if err != nil {
		fmt.Printf("could not create file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	// for debug
	nodes, ways, relations := 0, 0, 0
	snodes, sways, srelations := 0, 0, 0

	// ====================================
	// 3.recreate scanner and write GeoJSON
	// ====================================
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)
	scanner.SkipNodes = true
	scanner.SkipWays = true

	// write "FeatureCollection" record
	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	bfirst, bmp := true, true
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Relation:
			// target Tag
			if e.Tags.Find(tagname) == tagval {
				//
				// この段階で出力要不要の判定が必要
				//
				geojson := ""

				srelations++
				if bfirst {
					bfirst = false
				} else {
					geojson += ",\n"
				}

				switch e.Tags.Find("type") {
				case "multipolygon":
					// continue
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiPolygon\","
					bmp = true
				case "site":
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiLineString\","
					bmp = false
				}

				i := 0
				lopen := false
				// ******************************
				// start coordinates
				// このforをbreakすればGeoJSONはクリアされる
				// ******************************
				geojson += "\"coordinates\":["
				for _, v := range e.Members {
					// element kind "Way" only processing
					if v.Type == "way" {
						// get "Way" information from mrway( dictionary )
						if way, flg := mrway[int(v.Ref)]; flg {
							// split way information
							wayelm := strings.Split(way, "/")
							// mrway delimter "/".
							// wayelm array format.
							//  [0]: relation type("multipolygon" ,"site")
							//  [1]: element type( "way" ,"node" )
							//  [2]: role( "outer","inner","entrance","perimeter","label","" )
							//  [3]: coordinates
							//  [4]: first coordinate
							//  [5]: last coordinate
							//  [6]: open/close area( "open"/"close" )
							if i == 0 {
								if wayelm[6] == "open" {
									file.WriteString("[[")
								} else {
									file.WriteString("[")
								}
								i++
							} else {
								if wayelm[6] == "close" && wayelm[2] == "outer" {
									file.WriteString("],[")
								} else {
									file.WriteString(",")
								}
							}
							// file.WriteString(way)
							if wayelm[6] == "close" {
								file.WriteString("[" + wayelm[3] + "]")
							} else {
								if lopen {
									file.WriteString(wayelm[3])
								} else if i == 0 {
									file.WriteString(wayelm[3])
								} else {
									file.WriteString("[" + wayelm[3])
									lopen = true
								}
							}
						}
					}
				}
				// ******************************
				// close coordinates
				// ******************************
				file.WriteString("}")

				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("name"), "\\") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1)))
				} else {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name")))
				}

				// For debug
				fmt.Println("Relation Type:", e.Tags.Find("type"))
			}
			relations++
		}

	}
	scanner.Close()
	// FeatureCollection終端を出力
	file.WriteString("]}\n")

	// result
	end := time.Now()
	fmt.Println("Start:", start, "\nEnd:", end, "\nElapsed:", end.Sub(start))
	fmt.Println("nodes[", nodes, "] ways[", ways, "] relations[", relations, "]\nsnodes[", snodes, "] sways[", sways, "] srelations[", srelations, "]")
}
