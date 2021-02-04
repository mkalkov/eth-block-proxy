package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const applicationJSON = "application/json"

type proxyServer struct {
	gateway      string
	cache        *blockCache
	fetchCounter int
}

func newProxyServer(gateway string, cache *blockCache) proxyServer {
	return proxyServer{
		gateway:      gateway,
		cache:        cache,
		fetchCounter: 0,
	}
}

func (ps *proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println()
	logger.Println("Serving a request for", r.URL.Path)

	blockNr, _, err := parseURL(r.URL)
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Println("Looking up block", blockNr, "in cache")
	block, err := ps.cache.getBlockByNumber(blockNr)
	if err != nil {
		logger.Println(err)
		block, err = ps.fetchBlock(blockNr)
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if shallCache(blockNr) {
			ps.cache.putOrUpdate(blockNr, block)
		} else {
			logger.Println("Block", blockNr, "will not be cached")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", applicationJSON)
	written, err := w.Write([]byte(block))
	if err != nil {
		logger.Println("Error replying to client")
		return
	}
	logger.Println("Sent", written, "bytes to client")

	// TODO: parse block X in order to return only transaction Y by its index or its hash
}

// Expect URLs like /block/X/txs/Y
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
	}

	return blockPattern, path[4], nil
}

// See documentation at https://eth.wiki/json-rpc/API#eth_getblockbynumber
// Example call: curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0xb34f16", true],"id":1}'
func (ps *proxyServer) fetchBlock(blockNr string) (string, error) {
	ps.fetchCounter++
	blockIDString := "latest"
	if blockNr != "latest" {
		blockInt, _ := strconv.Atoi(blockNr)
		blockIDString = fmt.Sprintf("%#x", blockInt)
	}
	rpcString := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["%s", true],"id":%d}`, blockIDString, ps.fetchCounter)
	logger.Println("Performing JSON-RPC:", rpcString)

	// TODO: Configure timeouts
	resp, err := http.Post(ps.gateway, applicationJSON, strings.NewReader(rpcString))
	if err != nil {
		resp.Body.Close()
		return "", err
	}

	fetchedBlock, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = resp.Body.Close()
	if err != nil {
		return "", err
	}
	logger.Println("Fetched", len(fetchedBlock), "bytes of", blockNr, "block")

	return string(fetchedBlock), err
}
