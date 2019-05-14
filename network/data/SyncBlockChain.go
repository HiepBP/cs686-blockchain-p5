package data

import (
	"sync"
	"time"

	bc "../../blockchain"
	"../../mpt"
)

var blockReward = 5

//SyncBlockChain contain the BlockChain and sync support
type SyncBlockChain struct {
	bc  bc.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: bc.NewBlockChain()}
}

//GetLatestBlocks returns the list of blocks of height "BlockChain.length".
func (sbc *SyncBlockChain) GetLatestBlocks() []bc.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

//GetParentBlock takes a block as the parameter, and returns its parent block
func (sbc *SyncBlockChain) GetParentBlock(block bc.Block) (bc.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

//Get will return a blocks at specific height
func (sbc *SyncBlockChain) Get(height int32) ([]bc.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	//return sbc.bc.Get(height)
	blocks := sbc.bc.Get(height)
	if len(blocks) != 0 {
		return blocks, true
	}
	return blocks, false
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (bc.Block, bool) {
	blocks, available := sbc.Get(height)
	if available {
		for _, block := range blocks {
			if block.Header.Hash == hash {
				return block, true
			}
		}
	}
	return bc.Block{}, false
}

func (sbc *SyncBlockChain) Insert(block bc.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

//CheckParentHash will check if parent block available base on insertBlock Header
func (sbc *SyncBlockChain) CheckParentHash(insertBlock bc.Block) bool {
	_, available := sbc.GetBlock(insertBlock.Header.Height-1, insertBlock.Header.ParentHash)
	return available
}

//UpdateEntireBlockChain will add the downloaded bc into current one
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJSON string) {
	sbc.mux.Lock()
	sbc.bc, _ = bc.DecodeJSONToBlockChain(blockChainJSON)
	sbc.mux.Unlock()
}

//BlockChainToJson convert from BlockChain to Json string
func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJSON()
}

//GenBlock will create new block from MPT
func (sbc *SyncBlockChain) GenBlock(mpt mpt.MerklePatriciaTrie, nonce string,
	parentHash string, minedBy string, accountTrie *mpt.MerklePatriciaTrie) bc.Block {
	height := sbc.bc.Length

	var result bc.Block
	// currentTime, _ := strconv.ParseInt(time.Now().Format(time.RFC850), 10, 64)
	currentTime := time.Now().Unix()
	//Add reward
	minerAccountJSON, _ := accountTrie.Get(minedBy)
	minerAccount, _ := bc.DecodeAccountFromJSON(minerAccountJSON)
	minerAccount.Balance = minerAccount.Balance + blockReward
	minerAccountJSON, _ = minerAccount.EncodeToJSON()
	accountTrie.Insert(minedBy, minerAccountJSON)
	if parentHash != "" {
		result = bc.Initial(height+1, currentTime, parentHash, mpt, nonce, accountTrie.Root, minedBy, blockReward)
	} else {
		result = bc.Initial(height+1, currentTime, "", mpt, nonce, accountTrie.Root, minedBy, blockReward)
	}
	return result
}

//Show will return a string of BlockChain
func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}
