package bc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"golang.org/x/crypto/sha3"
)

//BlockChain contain a map(which maps to block height to a list of blocks)
//and Length equals to the highest block height
type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

//BlockChainJSON is a json class for BlockChain
type BlockChainJSON struct {
	Chain []BlockJSON
}

//BlockJSON is a json class for array of Block
type BlockJSON struct {
	Elements map[string]string `json:" "`
}

//NewBlockChain is Constructor for BlockChain
func NewBlockChain() BlockChain {
	Chain := make(map[int32][]Block)
	Length := int32(0)
	return BlockChain{Chain, Length}
}

//Get function take a height then return the lists of blocks stored in that height
//or None if the height doesn't exist
func (blockChain *BlockChain) Get(height int32) []Block {
	if block, available := blockChain.Chain[height]; available {
		return block
	}
	return nil
}

//Insert function takes a block then insert it into BlockChain.chain
//If the list has already contained that block's hash,
//ignore it because we don't store duplicate blocks; if not, insert the block into the list.
//QUESTION: do we need to check parentHash in previous height before Insert
//QUESTION: if block chain at height 4, can we insert a new block at height 1 or 2?
func (blockChain *BlockChain) Insert(newBlock Block) {
	if listBlock, ok := blockChain.Chain[newBlock.Header.Height]; ok { //Already has that height
		if !containHash(listBlock, newBlock) {
			listBlock = append(listBlock, newBlock)
			blockChain.Chain[newBlock.Header.Height] = listBlock
		}
	} else { //New height
		blockChain.Chain[newBlock.Header.Height] = []Block{newBlock}
		blockChain.Length = newBlock.Header.Height
	}
}

//EncodeToJSON function encode the blockchain instance into a json string
func (blockChain *BlockChain) EncodeToJSON() (string, error) {
	var result string
	var elements []map[string]interface{}
	var element map[string]interface{}
	for _, blockHeight := range blockChain.Chain {
		for _, block := range blockHeight {
			element = make(map[string]interface{})
			jsonBlock, err := block.EncodeToJSON()
			if err != nil {
				return result, err
			}
			//Pass jsonBlock back into map of key value in json
			err = json.Unmarshal([]byte(jsonBlock), &element)
			if err != nil {
				return result, err
			}
			elements = append(elements, element)
		}
	}
	jsonByte, _ := json.Marshal(elements)
	result = string(jsonByte)
	fmt.Println(string(jsonByte))
	return result, nil
}

//DecodeJSONToBlockChain function call by blockchain instance and take json string
//decode that string back to blockchain instance and copy everthing to the current blockchain
func DecodeJSONToBlockChain(jsonString string) (BlockChain, error) {
	result := NewBlockChain()
	var elements []map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &elements)
	for _, element := range elements {
		jsonElement, _ := json.Marshal(element)
		block, _ := DecodeFromJSON(string(jsonElement))
		result.Insert(block)
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func containHash(listBlock []Block, block Block) bool {
	for _, currBlock := range listBlock {
		if currBlock.Header.Hash == block.Header.Hash {
			return true
		}
	}
	return false
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

//GetLatestBLocks returns the list of blocks of height "BlockChain.length".
func (blockChain *BlockChain) GetLatestBlocks() []Block {
	return blockChain.Get(blockChain.Length)
}

//GetParentBlock takes a block as the parameter, and returns its parent block
func (blockChain *BlockChain) GetParentBlock(block Block) (Block, bool) {
	blocks := blockChain.Get(block.Header.Height - 1)
	for _, ancestorBlock := range blocks {
		if ancestorBlock.Header.Hash == block.Header.ParentHash {
			return ancestorBlock, true
		}
	}
	return Block{}, false
}
