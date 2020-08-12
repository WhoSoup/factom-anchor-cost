package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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
}

type btcinput struct {
	PrevOut btcout `json:"prev_out"`
}

type btcout struct {
	Spent bool   `json:"spent"`
	Value uint64 `json:"value"`
}

func (b *BTC) Get(txid string) (uint64, error) {
	body, err := b.call(txid)
	if err != nil {
		return 0, err
	}

	res := btcresp{}
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, err
	}

	var input uint64
	for _, in := range res.Inputs {
		if in.PrevOut.Spent {
			input += in.PrevOut.Value
		}
	}
	for _, out := range res.Out {
		if out.Spent {
			input -= out.Value
		}
	}

	if input < 0 {
		return 0, errors.New("negative spend")
	}
	return input, nil
}
