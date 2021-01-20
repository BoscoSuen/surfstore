package surfstore

import (
    "crypto/sha256"
    "encoding/hex"
)

type BlockStore struct {
    BlockMap map[string]Block
}

/**
* Retrieves a block indexed by hash value h.
*/
func (bs *BlockStore) GetBlock(blockHash string, blockData *Block) error {
    *blockData = bs.BlockMap[blockHash]
    return nil
}

/**
* Stores block b in the key-value store, indexed by hash value h.
* For each block, a hash value is generated using the SHA-256 hash function.
*/
func (bs *BlockStore) PutBlock(block Block, succ *bool) error {
    hash := sha256.New()
    hash.Write(block.BlockData)
    hashBytes := hash.Sum(nil)
    hashCode := hex.EncodeToString(hashBytes)
    bs.BlockMap[hashCode] = block
    return nil
}

/**
* Given a list of hashes “in”, returns a list containing the subset 
* of in that are stored in the key-value store.
*/
func (bs *BlockStore) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
    for _, blockHash := range blockHashesIn {
        if _, ok := bs.BlockMap[blockHash]; ok {
            *blockHashesOut = append(*blockHashesOut, blockHash)
        }
    }
    return nil
}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)
