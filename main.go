package main

import (
	"flag"
	"fmt"

	"github.com/FactomProject/factom"
)

var eth *Ethplorer

func main() {
	server := flag.String("s", "localhost:8088", "The location of the factomd api")
	ethapi := flag.String("eth", "", "The API key for ethplorer.io")
	flag.Parse()

	if *ethapi == "" {
		panic("no eth api key provided")
	}

	eth = NewEthplorer(*ethapi)
	factom.SetFactomdServer(*server)

	doHeight(230000)
}

func doHeight(h int64) error {
	anchor, err := factom.GetAnchorsByHeight(h)
	if err != nil {
		panic(err)
	}
	fmt.Println(anchor)

	e, err := eth.Get(anchor.Ethereum.TxID)
	if err != nil {
		return err
	}

	fmt.Printf("paid %v", e)
	return nil
}
