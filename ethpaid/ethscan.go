package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/ratelimit"
)

// https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=0x1e2910a262b1008d0616a0beb24c1a491d78771baa54a33e66065e03b1f46bc1&apikey=YourApiKeyToken
// https://api.etherscan.io/api?module=proxy&action=eth_getTransactionReceipt&txhash=0x1e2910a262b1008d0616a0beb24c1a491d78771baa54a33e66065e03b1f46bc1&apikey=YourApiKeyToken

const ETH_URL = "https://api.etherscan.io/api?module=proxy&action=eth_%s&txhash=%s&apikey=%s"
const ETH2_URL = "https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=false&apikey=%s"
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

func (e *Ethscan) call(url string) ([]byte, error) {
	e.limit.Take()

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

func (e *Ethscan) wrap(url string) (map[string]interface{}, error) {
	body, err := e.call(url)
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

func (e *Ethscan) Get(txid string) (string, error) {
	url := fmt.Sprintf(ETH_URL, "getTransactionByHash", txid, e.key)
	res, err := e.wrap(url)
	if err != nil {
		return "", err
	}

	number := res["blockNumber"]

	url = fmt.Sprintf(ETH2_URL, number, e.key)
	res, err = e.wrap(url)
	if err != nil {
		return "", err
	}

	unixts, err := ethconv(res["timestamp"])
	if err != nil {
		return "", err
	}

	return time.Unix(int64(unixts), 0).Format("2006-01-02 15:04"), nil
}

func ethconv(num interface{}) (uint64, error) {
	return strconv.ParseUint((strings.Replace(fmt.Sprintf("%v", num), "0x", "", 1)), 16, 64)
}
