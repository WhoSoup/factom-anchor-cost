package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func p(err error) {
	if err != nil {
		panic(err)
	}
}

func loadBlockTimes() map[int]time.Time {
	blocktimes := make(map[int]time.Time)

	btfile, err := os.Open("blocktime.json")
	p(err)

	btdata, err := ioutil.ReadAll(btfile)
	p(err)

	err = json.Unmarshal(btdata, &blocktimes)
	p(err)

	return blocktimes
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

func loadPrices(fname string) map[time.Time]float64 {

	f, err := os.Open(fname)
	p(err)
	defer f.Close()
	/*
	   Timestamps are UTC timezone,https://www.CryptoDataDownload.com
	   Date,Symbol,Open,High,Low,Close,Volume BTC,Volume USD
	   2020-08-10 09-PM,BTCUSD,11863.32,11869.26,11800.06,11834.33,571.75,6759675.81
	   0                  1      2        3        4         5        6     7
	*/
	res := make(map[time.Time]float64)
	sc := bufio.NewScanner(f)
	skip := 2
	for sc.Scan() {
		if skip > 0 {
			skip--
			continue
		}

		tokens := strings.Split(sc.Text(), ",")

		t, err := time.Parse("2006-01-02", tokens[0])
		p(err)

		high, err := strconv.ParseFloat(tokens[3], 64)
		p(err)
		low, err := strconv.ParseFloat(tokens[4], 64)

		res[t] = (high + low) / 2
	}
	return res

}

func main() {
	btcPrice := loadPrices("Coinbase_BTCUSD_d.csv")
	ethPrice := loadPrices("Coinbase_ETHUSD_d.csv")
	blocktimes := loadBlockTimes()
	btc := loadCosts("bitcoin.txt")
	eth := loadCosts("ethereum.txt")

	stitch("btc-stitch.txt", "BTC", btcPrice, blocktimes, btc)
	stitch("eth-stitch.txt", "ETH", ethPrice, blocktimes, eth)
}

func stitch(out, symbol string, prices map[time.Time]float64, blocktimes map[int]time.Time, costs []Fee) {
	f, err := os.Create(out)
	p(err)
	defer f.Close()

	fmt.Fprintln(f, "Block,Time,Price,Fee,FeeUSD,Cumulative,CumulativeUSD")
	cum := 0.0
	cumusd := 0.0

	for _, c := range costs {
		t := blocktimes[c.Height]
		hour := t.Add(-time.Duration(t.Minute()) * time.Minute)
		hour = hour.Add(-time.Duration(t.Hour()) * time.Hour)

		price := prices[hour]

		val := c.Fee * price
		cum += c.Fee
		cumusd += val

		fmt.Fprintf(f, "%d,%s,%f,%f,%f,%f,%f\n", c.Height, t.Format("2006-01-02 15:04"), price, c.Fee, val, cum, cumusd)
	}
}
