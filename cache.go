package main

import (
	"errors"
	"math"
	"strconv"
	"sync"
)

// BlockCache stores blocks in a map ordered by BlockID
type BlockCache struct {
	mu        sync.Mutex
	entries   blockByNumberMap
	callCount uint32
	capacity  uint32
}

type blockByNumberMap map[BlockID]*struct {
	block    string
	lastUsed uint32
}

// NewBlockCache creates and initializes a new cache with a given capacity
func NewBlockCache(capacity uint32) *BlockCache {
	// TODO: Fetch latest block number every 15s to know what can be cached
	// https://eth.wiki/json-rpc/API#eth_blocknumber
	// curl https://cloudflare-eth.com --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
	return &BlockCache{
		entries:   make(blockByNumberMap, capacity),
		callCount: 0,
		capacity:  capacity,
	}
}

// Get returns cached block or an error otherwise
func (cache *BlockCache) Get(blockID BlockID) (string, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if entry, ok := cache.entries[blockID]; ok {
		cache.callCount++
		entry.lastUsed = cache.callCount
		return entry.block, nil
	}
	return "", errors.New("Block " + string(blockID) + " is not cached")
}

// PutOrUpdate caches a block or just updates its lastUsed property
func (cache *BlockCache) PutOrUpdate(blockID BlockID, block string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.callCount++
	if blockEntry, ok := cache.entries[blockID]; ok {
		logger.Println("Block", blockID, "is already cached")
		blockEntry.lastUsed = cache.callCount
	} else {
		cache.expungeOldEntries()
		logger.Println("Block", blockID, "will be cached")
		cache.entries[blockID] = &struct {
			block    string
			lastUsed uint32
		}{
			block:    block,
			lastUsed: cache.callCount,
		}
	}
}

func (cache *BlockCache) expungeOldEntries() {
	for uint32(len(cache.entries)) >= cache.capacity {
		var blockID BlockID = ""
		var lastUsed uint32 = math.MaxUint32
		for k, v := range cache.entries {
			if v.lastUsed < lastUsed {
				blockID = k
				lastUsed = v.lastUsed
			}
		}
		logger.Println("Removing", blockID, "from cache")
		delete(cache.entries, blockID)
	}
}

// ShallCache tells whether a block with a given ID shall be cached or not
func ShallCache(blockID BlockID) bool {
	_, err := strconv.Atoi(string(blockID))
	return err == nil
}
