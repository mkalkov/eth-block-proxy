package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mkalkov/eth-block-proxy/ethproxy"
)

const defaultEthereumGateway = "https://cloudflare-eth.com/"
const defaultProxyPort = 8000
const defaultCacheCapacity = 10

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// TODO: Receive configuration through flags

	cache := ethproxy.NewBlockCache(defaultCacheCapacity)

	// TODO: Configure timeouts
	proxyServer := ethproxy.NewProxyServer(defaultEthereumGateway, cache)

	log.Println("Ready to proxy calls to", defaultEthereumGateway)
	http.ListenAndServe(fmt.Sprintf(":%d", defaultProxyPort), proxyServer)
}
