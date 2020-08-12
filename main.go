package main

import (
	"flag"
	"fmt"

	"github.com/FactomProject/factom"
)

var eth *Ethscan

func main() {
	server := flag.String("s", "localhost:8088", "The location of the factomd api")
	ethapi := flag.String("eth", "", "The API key for etherscan.io")
	flag.Parse()

	if *ethapi == "" {
		panic("no eth api key provided")
	}

	eth = NewEthscan(*ethapi)
	factom.SetFactomdServer(*server)

	if err := doHeight(230000); err != nil {
		fmt.Println("!!ERROR!!", err)
	}

}

func doHeight(h int64) error {
	anchor, err := factom.GetAnchorsByHeight(h)
	if err != nil {
		panic(err)
	}
	fmt.Println(anchor)

	price, used, err := eth.Get(anchor.Ethereum.TxID)
	if err != nil {
		return err
	}

	fmt.Println("gasPrice", price, "gwei", "Used", used)
	return nil
}
