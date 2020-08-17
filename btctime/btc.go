package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/ratelimit"
)

const BTC_URL = "https://blockchain.info/rawtx/%s"
const BTC_LIMIT = 5

type BTC struct {
	limit ratelimit.Limiter
}

func NewBTC() *BTC {
	b := new(BTC)
	b.limit = ratelimit.New(BTC_LIMIT)
	return b
}

func (b *BTC) call(hash string) ([]byte, error) {
	b.limit.Take()

	url := fmt.Sprintf(BTC_URL, hash)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	return ioutil.ReadAll(resp.Body)
}

type btcresp struct {
	Inputs []btcinput `json:"inputs"`
	Out    []btcout   `json:"out"`
	Time   int64      `json:"time"`
}

type btcinput struct {
	PrevOut btcout `json:"prev_out"`
}

type btcout struct {
	Spent bool   `json:"spent"`
	Value uint64 `json:"value"`
}

func (b *BTC) Get(txid string) (string, error) {
	body, err := b.call(txid)
	if err != nil {
		return "", err
	}

	res := btcresp{}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	return time.Unix(res.Time, 0).Format("2006-01-02 15:04"), nil
}
