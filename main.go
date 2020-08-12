package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FactomProject/factom"
)

var eth *Ethscan
var btc *BTC

func main() {
	server := flag.String("s", "localhost:8088", "The location of the factomd api")
	ethapi := flag.String("eth", "", "The API key for etherscan.io")
	startS := flag.Int64("start", 0, "Start height")
	endS := flag.Int64("end", 0, "End height")
	flag.Parse()

	if *ethapi == "" {
		panic("no eth api key provided")
	}

	eth = NewEthscan(*ethapi)
	btc = NewBTC()
	factom.SetFactomdServer(*server)

	start := *startS
	end := *endS

	if start < 0 {
		start = 0
	}

	if end < start || end <= 0 {
		end = -1
	}

	ethf, err := os.Create("eth.txt")
	if err != nil {
		panic(err)
	}
	defer ethf.Close()
	fmt.Fprintf(ethf, "Height, TxID, Eth Paid\n")

	btcf, err := os.Create("btc.txt")
	if err != nil {
		panic(err)
	}
	defer btcf.Close()
	fmt.Fprintf(btcf, "Height, TxID, BTC Fee\n")

	for i := start; ; i++ {
		if end > 0 && i > end {
			break
		}
		anchor, err := factom.GetAnchorsByHeight(i)
		if err != nil {
			fmt.Println("ERROR", i, err)
			break
		}

		if anchor.Bitcoin != nil {
			if spent, err := doBTC(anchor.Bitcoin.TransactionHash); err != nil {
				fmt.Println("ERROR", i, err)
				break
			} else {
				fmt.Fprintf(btcf, "%d, %s, %.9f\n", i, anchor.Bitcoin.TransactionHash, spent)
			}
		}

		if anchor.Ethereum != nil {
			if spent, err := doEth(anchor.Ethereum.TxID); err != nil {
				fmt.Println("ERROR", i, err)
				break
			} else {
				fmt.Fprintf(ethf, "%d, %s, %.9f\n", i, anchor.Ethereum.TxID, spent)
			}

		}
		fmt.Println("height", i, "done")
	}
}

func doEth(txid string) (float64, error) {
	price, used, err := eth.Get(txid)
	if err != nil {
		return 0, err
	}

	eth := float64(used*price) / 1e9
	//fmt.Println("gasPrice", price, "gwei", "Used", used, eth)
	return eth, nil
}

func doBTC(txid string) (float64, error) {
	btc, err := btc.Get(txid)
	if err != nil {
		return 0, err
	}
	return float64(btc) / 1e8, nil
}
