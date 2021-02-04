package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultGateway = "https://cloudflare-eth.com/"
const defaultProxyPort = 8000

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)
var gatewayCallCounter = 1

func main() {
	// TODO: send port number and gateway as optional parameters, print out gateway at startup

	// TODO: Fetch latest block number every 15s to know what can be cached
	// https://eth.wiki/json-rpc/API#eth_blocknumber
	// curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

	// TODO: Configure timeouts
	proxyServer := ProxyServer{defaultGateway}
	http.ListenAndServe(fmt.Sprintf(":%d", defaultProxyPort), proxyServer)
}
