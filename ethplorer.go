package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.uber.org/ratelimit"
)

const ETH_URL = "https://api.ethplorer.io/"
const ETH_LIMIT = 10

type Ethplorer struct {
	key   string
	limit ratelimit.Limiter
}

func NewEthplorer(key string) *Ethplorer {
	e := new(Ethplorer)
	e.key = key
	e.limit = ratelimit.New(ETH_LIMIT)
	return e
}

func (e *Ethplorer) Get(txid string) (interface{}, error) {
	e.limit.Take()

	url := fmt.Sprintf("%sgetTxInfo/%s?apiKey=%s", ETH_URL, txid, e.key)
	fmt.Println(url)

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if err, ok := res["error"]; ok {
		return nil, fmt.Errorf("%v", err)
	}

	return res, nil
}
