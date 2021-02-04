package main

import (
	"errors"
	"math"
	"strconv"
)

type blockCache struct {
	blockEntries blockByNumberMap
	callCounter  int
	capacity     int
}

type blockByNumberMap map[string]*struct {
	value    string
	lastUsed int
}

func newBlockCache(capacity int) blockCache {
	return blockCache{
		blockEntries: make(blockByNumberMap, capacity),
		callCounter:  0,
		capacity:     capacity,
	}
}

func (cache *blockCache) getBlockByNumber(blockNr string) (string, error) {
	if blockEntry, ok := cache.blockEntries[blockNr]; ok {
		cache.callCounter++
		blockEntry.lastUsed = cache.callCounter
		return blockEntry.value, nil
	}
	return "", errors.New("Block " + blockNr + " is not cached")
}

func (cache *blockCache) putOrUpdate(blockNr string, block string) {
	cache.callCounter++
	if blockEntry, ok := cache.blockEntries[blockNr]; ok {
		logger.Println("Block", blockNr, "is already cached")
		blockEntry.lastUsed = cache.callCounter
	} else {
		cache.expungeOldEntries()
		logger.Println("Block", blockNr, "will be cached")
		cache.blockEntries[blockNr] = &struct {
			value    string
			lastUsed int
		}{
			value:    block,
			lastUsed: cache.callCounter,
		}
	}
}

func (cache *blockCache) expungeOldEntries() {
	for len(cache.blockEntries) >= cache.capacity {
		var blockNr string = ""
		var lastUsed = math.MaxInt32
		for k, v := range cache.blockEntries {
			if v.lastUsed < lastUsed {
				blockNr = k
				lastUsed = v.lastUsed
			}
		}
		logger.Println("Removing", blockNr, "from cache")
		delete(cache.blockEntries, blockNr)
	}
}

func shallCache(blockNr string) bool {
	_, err := strconv.Atoi(blockNr)
	return err == nil
}
