package main

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func privateKeyToAddress(privateKey *big.Int) string {
	privKey, _ := btcec.PrivKeyFromBytes(privateKey.Bytes())
	pubKey := privKey.PubKey()
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	addr, _ := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	return addr.EncodeAddress()
}

func worker(start, end *big.Int, targetAddresses map[string]bool, result chan<- *big.Int, stop <-chan struct{}) {
	for i := new(big.Int).Set(start); i.Cmp(end) <= 0; i.Add(i, big.NewInt(1)) {
		select {
		case <-stop:
			return
		default:
			if targetAddresses[privateKeyToAddress(i)] {
				result <- new(big.Int).Set(i)
				return
			}
		}
	}
}

func searchForAddresses(startRange, endRange string, targetAddresses map[string]bool, numWorkers int) {
	start, _ := new(big.Int).SetString(startRange, 16)
	end, _ := new(big.Int).SetString(endRange, 16)

	result := make(chan *big.Int, 1)
	stop := make(chan struct{})
	var wg sync.WaitGroup

	rangeSize := new(big.Int).Sub(end, start)
	rangeSize.Add(rangeSize, big.NewInt(1))
	chunkSize := new(big.Int).Div(rangeSize, big.NewInt(int64(numWorkers)))

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerStart := new(big.Int).Mul(big.NewInt(int64(i)), chunkSize)
		workerStart.Add(workerStart, start)
		workerEnd := new(big.Int).Add(workerStart, chunkSize)
		if i == numWorkers-1 {
			workerEnd = end
		}
		go func() {
			defer wg.Done()
			worker(workerStart, workerEnd, targetAddresses, result, stop)
		}()
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	foundKey := <-result
	if foundKey != nil {
		close(stop)
		address := privateKeyToAddress(foundKey)
		fmt.Printf("Found Address: %s\nPrivate Key: %064x\n", address, foundKey)
	} else {
		fmt.Println("No matching address found in the given range.")
	}
}

func main() {
	startRange := "40000000000000000000" // change this to the start of the range you want to search
	endRange := "7fffffffffffffffffff"   // change this to the end of the range you want to search

	targetAddresses := map[string]bool{
		"": true, // change this to the address you want to search for
	}

	numCPU := runtime.NumCPU() // change this to the number of CPU cores you have available
	runtime.GOMAXPROCS(numCPU) // set the GOMAXPROCS environment variable to the number of CPU cores you have available

	searchForAddresses(startRange, endRange, targetAddresses, numCPU)
}
