package surfstore

import (
	context "context"
	"errors"
	"sync"
)

type BlockStore struct {
	suo      sync.Mutex
	BlockMap map[string]*Block
	UnimplementedBlockStoreServer
}

func (bs *BlockStore) GetBlock(ctx context.Context, blockHash *BlockHash) (*Block, error) {
	bs.suo.Lock()
	defer bs.suo.Unlock()
	res, flag := bs.BlockMap[blockHash.Hash]
	if flag {
		return res, nil
	}
	return nil, errors.New("get block failed")
}

func (bs *BlockStore) PutBlock(ctx context.Context, block *Block) (*Success, error) {
	bs.suo.Lock()
	defer bs.suo.Unlock()
	bs.BlockMap[GetBlockHashString(block.BlockData)] = block
	_, flag := bs.BlockMap[GetBlockHashString(block.BlockData)]
	if flag {
		return &Success{Flag: true}, nil
	}
	return nil, errors.New("put block failed")
}

// Given a list of hashes “in”, returns a list containing the
// subset of in that are stored in the key-value store
func (bs *BlockStore) HasBlocks(ctx context.Context, blockHashesIn *BlockHashes) (*BlockHashes, error) {
	bs.suo.Lock()
	defer bs.suo.Unlock()
	res := BlockHashes{}
	for _, v := range blockHashesIn.GetHashes() {
		_, flag := bs.BlockMap[v]
		if flag {
			res.Hashes = append(res.Hashes, v)
		}
	}
	return &res, nil
}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)

func NewBlockStore() *BlockStore {
	return &BlockStore{
		BlockMap: map[string]*Block{},
	}
}
