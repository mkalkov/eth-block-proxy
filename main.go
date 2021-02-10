package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultEthereumGateway = "https://cloudflare-eth.com/"
const defaultProxyPort = 8000
const defaultCacheCapacity = 10

// BlockID is either a natural number or a string like "latest"
type BlockID string

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func main() {
	// TODO: Received configuration through flags

	logger.Println("Ready to proxy calls to", defaultEthereumGateway)

	cache := NewBlockCache(defaultCacheCapacity)

	// TODO: Configure timeouts
	proxyServer := NewProxyServer(defaultEthereumGateway, cache)
	http.ListenAndServe(fmt.Sprintf(":%d", defaultProxyPort), proxyServer)
}
