package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

const urlPattern = "http://localhost:8000/block/%d/txs/%d"
const blockNr = 11751194
const txsNr = 7567
const attempts = 100

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		go request(&wg, i, fmt.Sprintf(urlPattern, blockNr+i, txsNr))
	}

	wg.Wait()
	log.Printf("Sent %d requests without any errors\n", attempts)
}

func request(wg *sync.WaitGroup, id int, url string) {
	wg.Add(1)
	defer wg.Done()
	log.Printf("Sending request %d to %s\n", id, url)
	_, err := http.Get(url)
	if err != nil {
		log.Fatalf("Got an error response for request %d to %s:\n\t%s\n", id, url, err)
	}
}
