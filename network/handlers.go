package network

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"../blockchain"
	"../mpt"
	"../utils"
	"../wallet"
	"../web/models"
	"./data"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
)

var (
	mineAddress        string
	blocksInTransit    = [][]byte{}
	memoryPool         = make(map[string]blockchain.Transaction)
	confirmedTx        = make(map[string]bool)
	SBC                data.SyncBlockChain
	Peers              data.PeerList
	FIRST_ADDR         = "http://localhost:8081"
	PEERS_DOWNLOAD_URL = FIRST_ADDR + "/peers"

	SELF_ADDR       = "http://localhost:8081"
	MAX_PEER        = 32
	ifStarted       = false
	heartBeatRevice = false
	keyString       = "000000"
	accounts        mpt.MerklePatriciaTrie
)

func StartBlockChain(w http.ResponseWriter, r *http.Request) {

	if ifStarted {
		fmt.Fprint(w, "Peer already started\n")
		return
	}
	ifStarted = true
	_, portStr, _ := net.SplitHostPort(r.Host)
	portNumber, _ := strconv.Atoi(portStr)
	initPeer(int32(portNumber))
	createGenesisBlock()
	fmt.Fprintf(w, "Blockchain init")
}

//Start will Register ID, download BlockChain, start HeartBeat
func StartMiner(w http.ResponseWriter, r *http.Request) {
	if ifStarted {
		fmt.Fprint(w, "Peer already started\n")
		return
	}
	ifStarted = true
	_, portStr, _ := net.SplitHostPort(r.Host)
	portNumber, _ := strconv.Atoi(portStr)
	initPeer(int32(portNumber))
	//Download BlockChain and peer list
	if "http://"+r.Host != FIRST_ADDR {
		SELF_ADDR = "http://" + r.Host
		Download()
	}
	privateKey, publicKey := wallet.GenerateKey()
	mineAddress = string(publicKey)
	newAccount := models.Account{
		PublicKey: publicKey,
	}
	accountJSON, _ := newAccount.EncodeToJSON()
	SendNewAccount(accountJSON)
	//Start sending HeartBeat
	go startHeartBeat()
	fmt.Fprintln(w, "You public key: ", hex.EncodeToString(publicKey))
	fmt.Fprintln(w, "Your private key: ", hex.EncodeToString(privateKey))
}

// Show will Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJSON, err := SBC.BlockChainToJson()
	if err != nil {
		//data.PrintError(err, "Upload")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	fmt.Fprint(w, blockChainJSON)
}

//UploadPeer will upload a PeerList to whoever called this method, return jsonStr
func UploadPeer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := "http://" + html.UnescapeString(vars["address"])
	id, _ := strconv.ParseInt(vars["id"], 10, 32)
	Peers.Add(address, int32(id))
	peerListJSON, err := Peers.PeerMapToJson()
	//Add self addr and id to peerList
	var peerList map[string]int32
	json.Unmarshal([]byte(peerListJSON), &peerList)
	peerList[SELF_ADDR] = Peers.GetSelfId()
	peerListByte, _ := json.Marshal(peerList)
	peerListJSON = string(peerListByte)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	fmt.Fprint(w, peerListJSON)
}

// UploadBlock will upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	height, _ := strconv.ParseInt(vars["height"], 10, 32)
	hash := vars["hash"]
	block, available := SBC.GetBlock(int32(height), hash)
	if available {
		blockJSON, _ := block.EncodeToJSON()
		fmt.Fprint(w, blockJSON)
		return
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

//HeartBeatReceive will add received peerList
//then add a new block
//then forward it to other peers if Hops>1
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive Heart Beat")
	heartBeatBuffer, _ := ioutil.ReadAll(r.Body)
	heartBeat, _ := data.DecodeHeartBeatFromJSON(string(heartBeatBuffer))
	//Add sender Address and ID
	if heartBeat.Addr != SELF_ADDR {
		Peers.Add(heartBeat.Addr, heartBeat.Id)
	}
	//Add received peerlist
	Peers.InjectPeerMapJson(heartBeat.PeerMapJson, SELF_ADDR)
	//Add new block
	if heartBeat.IfNewBlock {
		block, _ := blockchain.DecodeFromJSON(heartBeat.BlockJson)
		//Check PoW
		y := proofOfWork(block.Header.ParentHash, block.Header.Nonce, block.Value.Root())
		if y[:len(keyString)] != keyString {
			return
		}
		fmt.Println("VALID BLOCK")
		if block.Header.ParentHash != "" && !SBC.CheckParentHash(block) {
			askForBlock(block.Header.Height-1, block.Header.ParentHash)
		}
		SBC.Insert(block)
		heartBeatRevice = true
	}
	if heartBeat.Hops > 1 {
		heartBeat.Hops--
		forwardHeartBeat(heartBeat)
	}
}

func NewAccountReceive(w http.ResponseWriter, r *http.Request) {
	buffer, err := ioutil.ReadAll(r.Body)
	account, _ := models.DecodeAccountFromJSON(string(buffer))
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	accounts.Insert(string(account.PublicKey), "2000.0")
}

//Canonical will prints the current canonical chain, and chains of all forks if there are forks
func Canonical(w http.ResponseWriter, r *http.Request) {
	rs := ""
	blocks := SBC.GetLatestBlocks()
	for index, block := range blocks {
		currBlock := block
		rs += fmt.Sprintf("Chain #%d\n", index+1)
		for currBlock.Header.Hash != "" {
			rs += fmt.Sprintf("height=%d, timestamp=%d, hash=%s, parentHash=%s, size=%d\n",
				currBlock.Header.Height, currBlock.Header.Timestamp, currBlock.Header.Hash,
				currBlock.Header.ParentHash, currBlock.Header.Size)
			currBlock, _ = SBC.GetBlock(currBlock.Header.Height-1, currBlock.Header.ParentHash)
		}
		rs += "\n"
	}
	fmt.Fprint(w, rs)
}

func HandleTx(w http.ResponseWriter, r *http.Request) {
	buffer, _ := ioutil.ReadAll(r.Body)
	transaction, _ := blockchain.DecodeTransactionFromJSON(string(buffer))
	//Received tx is already confirmed
	if confirmedTx[transaction.ID] == true {
		return
	}
	memoryPool[transaction.ID] = transaction
	if len(memoryPool) >= 2 && len(mineAddress) > 0 {
		startTryingNonces()
	}
}

func initPeer(portNumber int32) {
	//Register
	registerLocal(portNumber)
	SBC = data.NewBlockChain()
}

func registerLocal(id int32) {
	Peers = data.NewPeerList(id, int32(MAX_PEER))
}

func Download() {
	//Download peer list from first node
	selfID := strconv.Itoa(int(Peers.GetSelfId()))
	response, err := http.Get(PEERS_DOWNLOAD_URL + "/" + html.EscapeString(SELF_ADDR[7:]) + "/" + selfID)
	if err != nil {
		fmt.Printf("The HTTP request failed with error: %s\n", err)
	} else {
		fmt.Println("DOWNLOAD FINISHED")
		peersJSONBuffer, _ := ioutil.ReadAll(response.Body)
		Peers.InjectPeerMapJson(string(peersJSONBuffer), SELF_ADDR)
	}

	//Download blockchain from radom peer in peerlist
	for addr := range Peers.Copy() {
		response, err := http.Get(addr + "/upload")
		if err == nil {
			bcJSONBuffer, _ := ioutil.ReadAll(response.Body)
			SBC.UpdateEntireBlockChain(string(bcJSONBuffer))
			break
		}
	}
}

func askForBlock(height int32, hash string) {
	peers := Peers.Copy()
	for addr := range peers {
		getBlockURL := addr + "/block/" + strconv.Itoa(int(height)) + "/" + hash
		fmt.Println("GET: " + getBlockURL)
		response, err := http.Get(getBlockURL)
		if err != nil {
			fmt.Println("Peer not available")
			Peers.Delete(addr)
		} else {
			blockBuffer, _ := ioutil.ReadAll(response.Body)
			parentBlock, _ := blockchain.DecodeFromJSON(string(blockBuffer))
			//Check if parent of parent block valid
			if parentBlock.Header.Height > 1 && !SBC.CheckParentHash(parentBlock) { //If not root Block and dont has its parent hash
				askForBlock(parentBlock.Header.Height-1, parentBlock.Header.ParentHash)
			}
			SBC.Insert(parentBlock)
			break
		}
	}
}

func forwardHeartBeat(heartBeatData data.HeartBeatData) {
	heartBeatJSON, _ := heartBeatData.EncodeToJSON()
	sendHeartBeat(heartBeatJSON)
}

func startHeartBeat() {
	for true {
		rand.Seed(time.Now().UnixNano())
		randomTime := rand.Intn(6) + 5
		time.Sleep(time.Duration(randomTime) * time.Minute)
		Peers.Rebalance()
		peersJSON, _ := Peers.PeerMapToJson()
		heartBeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), blockchain.Block{}, peersJSON, SELF_ADDR)
		heartBeatJSON, _ := heartBeat.EncodeToJSON()
		sendHeartBeat(heartBeatJSON)
	}
}

func sendHeartBeat(heartBeatJSON string) {
	peers := Peers.Copy()
	for addr := range peers {
		_, err := http.Post(addr+"/heartbeat/receive", "application/json", bytes.NewBuffer([]byte(heartBeatJSON)))
		if err != nil {
			fmt.Println("Peer not available")
			Peers.Delete(addr)
		}
	}
}

//startTryingNonces tell miner to mine
func startTryingNonces() {
	//Solve puzzle
	for true {
		y := ""
		//Get parents block
		parentBlocks := SBC.GetLatestBlocks()
		//Create MPT
		mpt, txIDs := createMPT()

		//TODO: Add transaction fee for the miner

		rand.Seed(time.Now().UnixNano())
		mpt.Insert("hello", strconv.Itoa(rand.Intn(100)))
		parentHash := ""
		if len(parentBlocks) != 0 {
			parentHash = parentBlocks[0].Header.Hash
		}
		//Generate nonce
		random, _ := utils.RandomHex(16)
		x := hex.EncodeToString(random)
		y = proofOfWork(parentHash, x, mpt.Root())
		//Verify nonce
		for y[:len(keyString)] != keyString && !heartBeatRevice {
			random, _ = utils.RandomHex(16)
			x = hex.EncodeToString(random)
			y = proofOfWork(parentHash, x, mpt.Root())
		}
		if heartBeatRevice {
			heartBeatRevice = false
			fmt.Println("Receive Block")
			continue
		} else {
			fmt.Print("Found Block: ")
			fmt.Println(y)
			Peers.Rebalance()
			peersJSON, _ := Peers.PeerMapToJson()
			newBlock := SBC.GenBlock(mpt, x, parentHash)
			heartBeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), newBlock, peersJSON, SELF_ADDR)
			heartBeatJSON, _ := heartBeat.EncodeToJSON()
			sendHeartBeat(heartBeatJSON)
			deleteConfirmedTx(txIDs)
		}
		//If nonce found, create new block, send heart beat
		//If flag, continue;
	}
}

func createMPT() (mpt.MerklePatriciaTrie, []string) {
	mpt := mpt.MerklePatriciaTrie{}
	var transactionsID []string
	mpt.Initial()
	//Get  transaction from memoryPool, check if it not confirmed yet
	for id := range memoryPool {
		if confirmedTx[id] == true {
			delete(memoryPool, id)
			continue
		}
		transaction := memoryPool[id]
		transactionsID = append(transactionsID, transaction.ID)
		transactionJSON, _ := transaction.EncodeToJSON()
		mpt.Insert(transaction.ID, transactionJSON)
	}
	return mpt, transactionsID
}

func deleteConfirmedTx(ids []string) {
	for _, id := range ids {
		delete(memoryPool, id)
		confirmedTx[id] = true
	}
}

func proofOfWork(parentHash string, nonce string, rootHash string) string {
	str := parentHash + nonce + rootHash
	strByte := sha3.Sum256([]byte(str))
	y := hex.EncodeToString(strByte[:])
	return y
}

func createGenesisBlock() {
	currentTime := time.Now().Unix()
	mpt := mpt.MerklePatriciaTrie{}
	mpt.Initial()
	SBC.Insert(blockchain.Initial(0, currentTime, "", mpt, ""))
}

func validateBalance(publicKey string, amount float32) bool {
	account, _ := accounts.Get(publicKey)
	if account == "" {
		return false
	}
	accountBalance, _ := strconv.ParseFloat(account, 32)
	if float32(accountBalance) < amount {
		return false
	}
	return true
}

func validateTx(tx blockchain.SignedTransaction) bool {
	result, _ := wallet.ValidateTransaction(tx.Transaction, tx.Signature)
	if !result {
		fmt.Println("Signature not correct")
		return false
	}
	from := string(tx.Transaction.FromAddress)
	to := string(tx.Transaction.FromAddress)
	fromAccount, _ := accounts.Get(from)
	if fromAccount == "" {
		return false
	}
	fromBalance, _ := strconv.ParseFloat(fromAccount, 32)
	if float32(fromBalance) < tx.Transaction.Value {
		return false
	}
	toAccount, _ := accounts.Get(to)
	if toAccount == "" {
		accounts.Insert(toAccount, "0.0")
	}
	return true
}

func SendNewAccount(accountJson string) {
	peers := Peers.Copy()
	for addr := range peers {
		_, err := http.Post(addr+"/account/receive", "application/json", bytes.NewBuffer([]byte(accountJson)))
		if err != nil {
			fmt.Println("Peer not available")
			Peers.Delete(addr)
		}
	}
}
