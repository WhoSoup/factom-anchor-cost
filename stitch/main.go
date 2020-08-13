package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FactomProject/factom"
)

func main() {
	btcsum, _ := os.Open("ethereum.txt")
	//ethsum,_ := os.Open("ethereum.txt")

	sc := bufio.NewScanner(btcsum)

	dupl := make(map[string]bool)

	var sum float64
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		lines := strings.Split(sc.Text(), ",")

		if dupl[lines[1]] {
			continue
		}
		dupl[lines[1]] = true

		f, err := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sum += f
	}

	fmt.Println("Total:", sum, "BTC")

}

func main2() {
	/*	btcf, err := os.Open("btc.txt")
		if err != nil {
			panic(err)
		}*/

	blocktimes := make(map[int]time.Time)

	factom.SetFactomdServer("spoon:8088")

	for i := int64(1); i < 257893; i++ {
		db, _, err := factom.GetDBlockByHeight(i)
		if err != nil {
			log.Println(err)
		}

		t := time.Unix(int64(db.Header.Timestamp)*60, 0).UTC()
		blocktimes[db.Header.DBHeight] = t
		if i%10000 == 0 {
			fmt.Println(i)
		}
	}

	btime, err := os.Create("blocktime.json")
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(blocktimes)
	if err != nil {
		panic(err)
	}
	fmt.Println(btime.Write(b))
	btime.Close()
}
