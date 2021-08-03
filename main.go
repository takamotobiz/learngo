package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func main() {

	start := time.Now()

	f, err := os.Open("./japan-low.osm.pbf")
	//f, err := os.Open("./shikoku-low.osm.pbf")
	// f, err := os.Open("../../data/japan-low.osm.pbf")
	// f, err := os.Open("../../data/planet-low.osm.pbf")
	if err != nil {
		fmt.Printf("could not open file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	// ファイルを書き込み用にオープン (mode=0666)
	file, err := os.Create("./output.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// FeatureCollectionレコード（ヘッダー的なもの）を出力
	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	nodes, ways, relations := 0, 0, 0
	snodes, sways, srelations := 0, 0, 0
	cpu := runtime.NumCPU()

	scanner := osmpbf.New(context.Background(), f, cpu)
	defer scanner.Close()

	var endl bool
	for scanner.Scan() {

		switch e := scanner.Object().(type) {
		case *osm.Node:
			if e.Tags.Find("amenity") == "school" {
				snodes++
				// 最後のレコード出力時にはカンマを出力しない
				if endl {
					file.WriteString(",\n")
				} else {
					endl = true
				}
				// 要素情報の出力
				file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Point\",")
				file.WriteString(fmt.Sprintf("\"coordinates\":[%.7f,%.7f]}", e.Lon, e.Lat))
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
			}
			nodes++
		case *osm.Way:
			if e.Tags.Find("amenity") == "school" {
				sways++
				// 最後のレコード出力時にはカンマを出力しない
				if endl {
					file.WriteString(",\n")
				} else {
					endl = true
				}
				// 要素情報の出力
				if e.Polygon() {
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",")
					file.WriteString("\"coordinates\":[[")
				} else {
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"LineString\",")
					file.WriteString("\"coordinates\":[")
				}
				for i, v := range e.Nodes {
					if i > 0 {
						file.WriteString(",")
					}
					file.WriteString(fmt.Sprintf("[%.7f,%.7f]", v.Lon, v.Lat))
				}
				if e.Polygon() {
					file.WriteString("]]}")
				} else {
					file.WriteString("]}")
				}

				// file.WriteString(fmt.Sprintf("\"coordinates\":[%.7f,%.7f]}", e.Lon, e.Lat))
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
			}
			ways++
		case *osm.Relation:
			srelations++
			relations++
		}

	}
	// FeatureCollection終端を出力
	file.WriteString("]}\n")

	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner returned error: %v", err)
		os.Exit(1)
	}

	end := time.Now()

	fmt.Println("Start:", start)
	fmt.Println("End  :", end)
	fmt.Println("Elapsed:", end.Sub(start))

	fmt.Println("nodes:", nodes)
	fmt.Println("ways:", ways)
	fmt.Println("relations:", relations)

	fmt.Println("snodes:", snodes)
	fmt.Println("sways:", sways)
	fmt.Println("srelations:", srelations)

}
