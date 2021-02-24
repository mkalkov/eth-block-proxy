package ethproxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const applicationJSON = "application/json"

// ProxyServer fetches and caches Ethereum blocks
type ProxyServer struct {
	gateway      string
	cache        *BlockCache
	fetchCounter uint32
}

// NewProxyServer creates and initializes a new proxy server using provided gateway URL and block cache
func NewProxyServer(gatewayURL string, blockCache *BlockCache) *ProxyServer {
	return &ProxyServer{
		gateway:      gatewayURL,
		cache:        blockCache,
		fetchCounter: 0,
	}
}

func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println()
	log.Println("Serving a request for", r.URL.Path)

	blockID, _, err := parseURL(r.URL)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Looking up block", blockID, "in cache")
	block, err := ps.cache.Get(blockID)
	// use other return mechanism than err
	if err != nil {
		log.Println(err)
		block, err = ps.fetchBlock(blockID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if ShallCache(blockID) {
			ps.cache.PutOrUpdate(blockID, block)
		} else {
			log.Println("Block", blockID, "will not be cached")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", applicationJSON)
	written, err := w.Write([]byte(block))
	if err != nil {
		log.Println("Error replying to client")
		return
	}
	log.Println("Sent", written, "bytes to client")

	// TODO: parse block X in order to return only transaction Y by its index or its hash
}

// Expect URLs like /block/X/txs/Y
func parseURL(url *url.URL) (BlockID, string, error) {

	path := strings.Split(url.Path, "/")
	if len(path) != 5 || path[1] != "block" || path[2] == "" || path[3] != "txs" || path[4] == "" {
		return "", "", errors.New("Invalid request path")
	}

	blockPattern := path[2]
	if blockPattern != "latest" {
		blockNumber, err := strconv.Atoi(path[2])
		if err != nil || blockNumber < 0 {
			return "", "", errors.New("Invalid block number: " + path[2])
		}
	}

	return BlockID(blockPattern), path[4], nil
}

// See documentation at https://eth.wiki/json-rpc/API#eth_getblockbynumber
// Example call: curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0xb34f16", true],"id":1}'
func (ps *ProxyServer) fetchBlock(blockID BlockID) (string, error) {
	ps.fetchCounter++

	// TODO: Move this code to a new function with BlockID receiver
	blockIDString := "latest"
	if blockID != "latest" {
		blockInt, _ := strconv.Atoi(string(blockID))
		blockIDString = fmt.Sprintf("%#x", blockInt)
	}
	rpcString := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["%s", true],"id":%d}`, blockIDString, ps.fetchCounter)
	log.Println("Performing JSON-RPC:", rpcString)

	// TODO: Configure timeouts
	resp, err := http.Post(ps.gateway, applicationJSON, strings.NewReader(rpcString))
	if err != nil {
		resp.Body.Close() // use defer
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
	log.Println("Fetched", len(fetchedBlock), "bytes of", blockID, "block")

	return string(fetchedBlock), err
}
