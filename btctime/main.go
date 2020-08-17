package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func p(err error) {
	if err != nil {
		panic(err)
	}
}

type Fee struct {
	Height int
	Hash   string
	Fee    float64
}

func loadCosts(fname string) []Fee {
	f, err := os.Open(fname)
	p(err)
	defer f.Close()

	dupl := make(map[string]bool)

	var res []Fee
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}

		tokens := strings.Split(sc.Text(), ",")

		txid := tokens[1]
		if dupl[txid] {
			continue
		}
		dupl[txid] = true

		height, err := strconv.Atoi(strings.TrimSpace(tokens[0]))
		p(err)
		fee, err := strconv.ParseFloat(strings.TrimSpace(tokens[2]), 64)
		p(err)

		res = append(res, Fee{
			Height: height,
			Hash:   strings.TrimSpace(tokens[1]),
			Fee:    fee,
		})
	}

	return res
}

func main() {
	costs := loadCosts("bitcoin.txt")

	btc := NewBTC()

	out, err := os.Create("bitcoin-dates.txt")
	p(err)
	fmt.Fprintf(out, "Height,TxID,BtcPaid,TxDate\n")
	for i, f := range costs {
		t, err := btc.Get(f.Hash)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Fprintf(out, "%d,%s,%f,%s\n", f.Height, f.Hash, f.Fee, t)
		fmt.Println(i, "/", len(costs))

	}
}
