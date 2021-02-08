package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const urlPattern = "http://localhost:8000/block/%d/txs/%d"
const blockNr = 11751194
const txsNr = 7567
const attempts = 100

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func main() {
	c := make(chan string)
	for i := 0; i < attempts; i++ {
		go request(c, i, fmt.Sprintf(urlPattern, blockNr+i, txsNr))
	}
	for i := 0; i < attempts; i++ {
		<-c
	}
	logger.Printf("Sent %d requests without any errors\n", attempts)
}

func request(c chan string, id int, url string) {
	logger.Printf("Sending request %d to %s\n", id, url)
	_, err := http.Get(url)
	if err != nil {
		logger.Fatalf("Got an error response for request %d to %s:\n\t%s\n", id, url, err)
	}
	c <- ""
}
