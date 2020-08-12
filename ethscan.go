package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/ratelimit"
)

// https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=0x1e2910a262b1008d0616a0beb24c1a491d78771baa54a33e66065e03b1f46bc1&apikey=YourApiKeyToken
// https://api.etherscan.io/api?module=proxy&action=eth_getTransactionReceipt&txhash=0x1e2910a262b1008d0616a0beb24c1a491d78771baa54a33e66065e03b1f46bc1&apikey=YourApiKeyToken

const ETH_URL = "https://api.etherscan.io/api?module=proxy&action=eth_%s&txhash=%s&apikey=%s"
const ETH_LIMIT = 5

type Ethscan struct {
	key   string
	limit ratelimit.Limiter
}

func NewEthscan(key string) *Ethscan {
	e := new(Ethscan)
	e.key = key
	e.limit = ratelimit.New(ETH_LIMIT)
	return e
}

func (e *Ethscan) call(action, hash string) ([]byte, error) {
	e.limit.Take()

	url := fmt.Sprintf(ETH_URL, action, hash, e.key)

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

func (e *Ethscan) wrap(action, txid string) (map[string]interface{}, error) {
	body, err := e.call(action, txid)
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

	return res["result"].(map[string]interface{}), nil
}

func (e *Ethscan) Get(txid string) (uint64, uint64, error) {
	res, err := e.wrap("getTransactionByHash", txid)
	if err != nil {
		return 0, 0, err
	}

	price, err := ethconv(res["gasPrice"])
	if err != nil {
		return 0, 0, err
	}

	res, err = e.wrap("getTransactionReceipt", txid)
	if err != nil {
		return 0, 0, err
	}

	used, err := ethconv(res["gasUsed"])
	//fmt.Println(res["gasUsed"], used)
	if err != nil {
		return 0, 0, err
	}

	return price / 1e9, used, nil
}

func ethconv(num interface{}) (uint64, error) {
	return strconv.ParseUint((strings.Replace(fmt.Sprintf("%v", num), "0x", "", 1)), 16, 64)
}
