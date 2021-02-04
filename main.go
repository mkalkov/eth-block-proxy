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

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func main() {
	logger.Println("Ready to proxy calls to", defaultEthereumGateway)

	cache := newBlockCache(defaultCacheCapacity)

	// TODO: Fetch latest block number every 15s to know what can be cached
	// https://eth.wiki/json-rpc/API#eth_blocknumber
	// curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

	// TODO: Configure timeouts
	proxyServer := newProxyServer(defaultEthereumGateway, &cache)
	http.ListenAndServe(fmt.Sprintf(":%d", defaultProxyPort), &proxyServer)
}
