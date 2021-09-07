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
const debugfname string = "./debug.json"

// define Tag
const tagname string = "amenity"
const tagval string = "school"

func main() {

	// get start time
	start := time.Now()

	// ************************************
	// 0.create debug file
	// ************************************
	dfile, err := os.Create(debugfname)
	if err != nil {
		fmt.Printf("could not create debug file: %v", err)
		os.Exit(1)
	}
	defer dfile.Close()

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
	// fmt.Println("mdebug[", mdebug, "]")

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
				mrway[int(re.ID)] += fmt.Sprintf("/[%.7f,%.7f]", flon, flat)
				mrway[int(re.ID)] += fmt.Sprintf("/[%.7f,%.7f]", llon, llat)
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

	var bnext bool
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Relation:
			// target Tag
			if e.Tags.Find(tagname) == tagval {
				//
				// この段階で出力要不要の判定が必要
				//
				var geojson string

				srelations++
				if bnext {
					geojson += ",\n"
				} else {
					bnext = true
				}
				// ここはTypeで分岐
				if e.Tags.Find("type") == "multipolygon" {
					// MultiPolygon
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiPolygon\","
				} else {
					// site(MultiLineString)
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiLineString\","
				}
				geojson += "\"coordinates\":["

				var dflag bool
				if e.Tags.Find("name") == "廿日市市立宮内小学校" {
					fmt.Println("")
					dflag = true
				}

				var cntcoord int
				//var firstcoord string
				var outfile bool
				// ******************************
				// start coordinates
				// このforをbreakすればGeoJSONはクリアされる
				// ******************************
				for _, v := range e.Members {
					// element kind "Way" only processing
					if v.Type == "way" {
						// set output flag
						outfile = true
						// get "Way" information from mrway( dictionary )
						if way, flg := mrway[int(v.Ref)]; flg {
							if dflag {
								dfile.WriteString(way + "\n")
							}
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

							// if wayelm[2] == "inner" && wayelm[6] == "open" {
							// 	dfile.WriteString(way + "\n")
							// }

							if cntcoord > 0 {
								geojson += ","
							}
							// ここはTypeで分岐
							if wayelm[0] == "multipolygon" {
								// MultiPolygon
								if wayelm[6] == "close" {
									// ******************************************
									// outer/innerの判定処理が必要
									// ※続く面がという判定が必要
									// ******************************************
									// closed element
									if wayelm[2] == "outer" {
										geojson += "[[" + wayelm[3] + "]]"
									} else {
										geojson += "[" + wayelm[3] + "]"
									}
									cntcoord++
								} else {
									// open element
									// ******************************************
									// ここに、以下の処理を実装する。
									// ・最初のopenを見つけたらフラグon（bopen）
									// ・同時にバッファ辞書へ追加、keyが先頭座標、valが座標本体
									// ・最後のopenを見つけたらフラグoff（どうやって判定するか？）
									// ・最初のメンバの最終座標でkeyを引き当て座標点列を構成
									// ・引き当て終了したらバッファ辞書のメンバを削除
									// ・バッファがなくなったら処理終了
									// ・メンバが残っていてkey引き当て失敗したらそいの要素は破棄
									// ******************************************
									// if cntcoord == 0 {
									// 	geojson += "[["
									// 	firstcoord = wayelm[4]
									// 	cntcoord++
									// }
									// geojson += wayelm[3]
									// if firstcoord == wayelm[5] {
									// 	geojson += "]]"
									// 	cntcoord = 0
									// }
									outfile = false
									break
								}
							} else {
								// site(MultiLineString)
								geojson += "[" + wayelm[3] + "]"
								cntcoord++
							}
						}
					} else {
						break
					}
				}
				// ******************************
				// close coordinates
				// ******************************
				geojson += "]}"

				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("name"), "\\") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1))
				} else {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name"))
				}

				// write file
				if outfile {
					file.WriteString(geojson)
				}

				// For debug
				if dflag {
					dflag = false
				}
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
