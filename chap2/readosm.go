package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/qedus/osmpbf"
)

func main() {

	// 開始時刻の表示
	fmt.Println("Start:", time.Now())

	// ターゲットのosm.pbfをオープン
	//f, err := os.Open("/Users/takamotokeiji/data/osm.pbf/shikoku-latest.osm.pbf")
	f, err := os.Open("/Users/takamotokeiji/data/osm.pbf/japan-latest.osm.pbf")
	if err != nil {
		log.Fatal(err)
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

	d := osmpbf.NewDecoder(f)

	// use more memory from the start, it is faster
	d.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
	}

	var nc, wc, rc uint64
	var endl bool
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Node（点要素）の場合の処理
				if v.Tags["amenity"] == "school" {
					// 最後のレコード出力時にはカンマを出力しない
					if endl {
						file.WriteString(",\n")
					} else {
						endl = true
					}
					// 要素情報の出力
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Point\",")
					file.WriteString(fmt.Sprintf("\"coordinates\":[%.7f,%.7f]}", v.Lon, v.Lat))
					// 属性文字のエスケープ関連文字の訂正
					if strings.Contains(v.Tags["name"], "\\") {
						file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(v.Tags["name"], "\\", "", -1)))
					} else if strings.Contains(v.Tags["name"], "\n") {
						file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(v.Tags["name"], "\n", "", -1)))
					} else if strings.Contains(v.Tags["name"], "\"") {
						file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(v.Tags["name"], "\"", "　", -1)))
					} else {
						file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", v.Tags["name"]))
					}

				}

				// Process Node v.
				nc++
			case *osmpbf.Way:
				// Process Way v.
				wc++
			case *osmpbf.Relation:
				// Process Relation v.
				rc++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}

	// FeatureCollection終端を出力
	file.WriteString("]}\n")

	// 要素数の表示
	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)

	// 終了時刻の表示
	fmt.Println("End:", time.Now())
}
