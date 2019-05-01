package data

import (
	"sync"
	"time"

	"../../blockchain"
	"../../mpt"
)

//SyncBlockChain contain the BlockChain and sync support
type SyncBlockChain struct {
	bc  blockchain.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: blockchain.NewBlockChain()}
}

//GetLatestBlocks returns the list of blocks of height "BlockChain.length".
func (sbc *SyncBlockChain) GetLatestBlocks() []blockchain.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

//GetParentBlock takes a block as the parameter, and returns its parent block
func (sbc *SyncBlockChain) GetParentBlock(block blockchain.Block) (blockchain.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

//Get will return a blocks at specific height
func (sbc *SyncBlockChain) Get(height int32) ([]blockchain.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	//return sbc.bc.Get(height)
	blocks := sbc.bc.Get(height)
	if len(blocks) != 0 {
		return blocks, true
	}
	return blocks, false
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (blockchain.Block, bool) {
	blocks, available := sbc.Get(height)
	if available {
		for _, block := range blocks {
			if block.Header.Hash == hash {
				return block, true
			}
		}
	}
	return blockchain.Block{}, false
}

func (sbc *SyncBlockChain) Insert(block blockchain.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

//CheckParentHash will check if parent block available base on insertBlock Header
func (sbc *SyncBlockChain) CheckParentHash(insertBlock blockchain.Block) bool {
	_, available := sbc.GetBlock(insertBlock.Header.Height-1, insertBlock.Header.ParentHash)
	return available
}

//UpdateEntireBlockChain will add the downloaded bc into current one
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJSON string) {
	sbc.mux.Lock()
	sbc.bc, _ = blockchain.DecodeJSONToBlockChain(blockChainJSON)
	sbc.mux.Unlock()
}

//BlockChainToJson convert from BlockChain to Json string
func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJSON()
}

//GenBlock will create new blockchain from MPT
func (sbc *SyncBlockChain) GenBlock(mpt mpt.MerklePatriciaTrie, nonce string, parentHash string) blockchain.Block {
	height := sbc.bc.Length

	var result blockchain.Block
	// currentTime, _ := strconv.ParseInt(time.Now().Format(time.RFC850), 10, 64)
	currentTime := time.Now().Unix()
	if parentHash != "" {
		result = blockchain.Initial(height+1, currentTime, parentHash, mpt, nonce)
	} else {
		result = blockchain.Initial(height+1, currentTime, "", mpt, nonce)
	}
	return result
}

//Show will return a string of BlockChain
func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}
