package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ProxyServer for Ethereal getBlockByNumber calls
type ProxyServer struct{ gateway string }

// TODO: Return HTTP error in case of errors
func (ps ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("Serving a request for", r.URL.Path)

	block, txs, err := parseURL(r.URL)
	if err != nil {
		logger.Println(err)
		return
	}

	// TODO: implement LRU caching

	gatewayCallCounter++
	// https://eth.wiki/json-rpc/API#eth_getblockbynumber
	// curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0xb34f16", true],"id":1}'
	rpcString := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["%s", true],"id":%d}`, block, gatewayCallCounter)
	logger.Println("Fetching block by number:", rpcString)

	// TODO: Configure timeouts
	resp, err := http.Post(ps.gateway, "application/json", strings.NewReader(rpcString))
	if err != nil {
		logger.Println(err)
		return
	}
	defer resp.Body.Close()
	logger.Println("Received block", block)

	// TODO: parse block X in order to only return transaction Y by index or hash
	logger.Println("Returning the whole block instead of transaction", txs)
	written, err := io.Copy(w, resp.Body)
	logger.Println("Sent", written, "bytes to user")
}

// expect URLs like /block/X/txs/Y
func parseURL(url *url.URL) (block string, txs string, err error) {

	path := strings.Split(url.Path, "/")
	if len(path) != 5 || path[1] != "block" || path[2] == "" || path[3] != "txs" || path[4] == "" {
		return "", "", errors.New("Invalid request path")
	}

	blockPattern := path[2]
	if block != "latest" {
		blockNumber, err := strconv.Atoi(path[2])
		if err != nil || blockNumber < 0 {
			return "", "", errors.New("Invalid block number: " + path[2])
		}
		blockPattern = fmt.Sprintf("%#x", blockNumber)
	}

	return blockPattern, path[4], nil
}
