package ethproxy

import (
	"fmt"
	"testing"
)

type block struct {
	id  BlockID
	val string
}

func TestCache(t *testing.T) {
	tests := []struct {
		given    []block
		expected []block
	}{
		{ // empty cache
			[]block{},
			[]block{{"123", ""}},
		}, { // cache with one item
			[]block{{"123", "asdf"}},
			[]block{{"123", "asdf"}, {"456", ""}},
		}, { // cache with one item does not update value for already cached items
			[]block{{"123", "asdf"}, {"123", "asdf2"}},
			[]block{{"123", "asdf"}, {"456", ""}},
		}, { // cache with two items
			[]block{{"123", "asdf"}, {"456", "qwer"}},
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"789", ""}},
		}, { // cache with two items does not depend on order
			[]block{{"456", "qwer"}, {"123", "asdf"}},
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"789", ""}},
		}, { // cache with three items is not affected by repeated values
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"456", "qwer"}, {"789", "zxcv"}},
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"789", "zxcv"}, {"000", ""}},
		}, { // cache with four items expunges the least recently used item
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"789", "zxcv"}, {"000", "aaaa"}},
			[]block{{"123", ""}, {"456", "qwer"}, {"789", "zxcv"}, {"000", "aaaa"}, {"111", ""}},
		}, { // cache with four items expunges the least recently used item, and saves its new value
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"789", "zxcv"}, {"000", "aaaa"}, {"123", "asdf2"}},
			[]block{{"123", "asdf2"}, {"456", ""}, {"789", "zxcv"}, {"000", "aaaa"}, {"111", ""}},
		}, { // cache with four items expunges the least recently used item, taking repeats into account
			[]block{{"123", "asdf"}, {"456", "qwer"}, {"123", "asdf"}, {"789", "zxcv"}, {"000", "aaaa"}},
			[]block{{"123", "asdf"}, {"456", ""}, {"789", "zxcv"}, {"000", "aaaa"}, {"111", ""}},
		},
	}

	for i, test := range tests {
		cache := NewBlockCache(3)
		for _, givenBlock := range test.given {
			cache.PutOrUpdate(givenBlock.id, givenBlock.val)
		}

		for j, assert := range test.expected {
			cache.verifyGet(t, &assert, fmt.Sprintf("Test %d case %d", i, j), &test.given)
		}
	}
}

func TestCacheIsLruNotLfu(t *testing.T) {
	cache := NewBlockCache(3)

	givenBlockList1 := []block{
		{"111", "asdf"},
		{"222", "qwer"},
		{"333", "zxcv"},
	}
	for _, givenBlock := range givenBlockList1 {
		cache.PutOrUpdate(givenBlock.id, givenBlock.val)
	}
	cache.verifyGet(t, &block{"111", "asdf"}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"111", "asdf"}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"222", "qwer"}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"333", "zxcv"}, "Test that cache is LRU not LFU", &givenBlockList1)

	cache.PutOrUpdate("444", "ghgh")
	givenBlockList1 = append(givenBlockList1, block{"444", "ghgh"})
	cache.verifyGet(t, &block{"111", ""}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"222", "qwer"}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"333", "zxcv"}, "Test that cache is LRU not LFU", &givenBlockList1)
	cache.verifyGet(t, &block{"444", "ghgh"}, "Test that cache is LRU not LFU", &givenBlockList1)
}

func (cache *BlockCache) verifyGet(t *testing.T, expected *block, location string, given *[]block) {
	result, err := cache.Get(expected.id)
	if expected.val != result {
		t.Errorf(
			"%s : cached value was incorrect: given %s when %s is requested, expected '%s' to be returned but got '%s'",
			location, *given, expected.id, expected.val, result,
		)
	}
	if expected.val == "" && err == nil || expected.val != "" && err != nil {
		t.Error(
			location, ": an empty block shall always mean an error, and an error shall always mean an empty block",
		)
	}
}
