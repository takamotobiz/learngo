package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jneo8/osmpbfparser-go"
)

func main() {
	parser := osmpbfparser.New(
		osmpbfparser.Args{
			PBFFile:     "/Users/takamotokeiji/data/osm.pbf/japan-latest.osm.pbf",
			LevelDBPath: "/tmp/osmpbfparser",
			BatchSize:   10000,
		},
	)
	var nc, wc, rc uint64
	// ファイルを書き込み用にオープン (mode=0666)
	file, err := os.Create("./output.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	str := "Start:"
	str += time.Now().Format("2006-01-02 15:04:05")
	str += "\n"
	fmt.Println(str)

	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	var fi bool = true
	for emt := range parser.Iterator() {
		// rawJSON, err := emt.ToGeoJSON()
		// if err != nil {
		//     log.Fatal(err)
		// }
		// fmt.Println(string(rawJSON))

		tags := emt.GetTags()
		if nv, fl := tags["amenity"]; fl {
			if nv == "school" {

				if fi {
					fi = false
				} else {
					file.Write([]byte(",\n"))
				}

				rawJSON, err := emt.ToGeoJSON()
				if err != nil {
					log.Fatal(err)
				}
				file.Write([]byte(rawJSON))

				switch emt.Type {
				case 0: //Node
					nc++
				case 1: //Way
					wc++
				case 2: //Relation
					rc++
				}
			}
		}
	}
	file.WriteString("]}\n")

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)

	str1 := "End:"
	str1 += time.Now().Format("2006-01-02 15:04:05")
	str1 += "\n"
	fmt.Println(str1)

	if err := parser.Err(); err != nil {
		log.Fatal(err)
	}
}
