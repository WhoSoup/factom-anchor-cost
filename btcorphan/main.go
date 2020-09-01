package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func p(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	offsetS := flag.Int64("offset", 0, "Offset")
	flag.Parse()

	btc := NewBTC()

	out, err := os.Create("orphans.txt")
	p(err)
	fmt.Fprintf(out, "TxID,Height,KeyMR,TxDate\n")

	pos := *offsetS
	for {
		txs, err := btc.GetAddr("1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF", pos)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 1)
			continue
		}

		for _, tx := range txs {
			if len(tx.Out) != 2 {
				continue
			}
			data, err := hex.DecodeString(tx.Out[1].Script)
			if err != nil {
				log.Println(tx.Hash, err)
				continue
			}

			if len(data) < 40 {
				continue
			}

			if bytes.Equal(data[:4], []byte("j(Fa")) {
				data = data[4:]
			} else {
				log.Printf("Unknown script: %x", data)
				continue
			}

			height := binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...))
			keymr := data[6:]
			t := time.Unix(tx.Time, 0).Format("2006-01-02 15:04")

			fmt.Fprintf(out, "%s,%d,%064x,%s\n", tx.Hash, height, keymr, t)

		}

		pos += int64(len(txs))
		fmt.Println("done", pos)
		time.Sleep(time.Second * 20)
	}

}
