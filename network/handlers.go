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

	bc "../blockchain"
	"../mpt"
	"../utils"
	"../wallet"
	"./data"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
)

var (
	FIRST_ADDR         = "http://localhost:8081"
	PEERS_DOWNLOAD_URL = FIRST_ADDR + "/peers"
	SELF_ADDR          = "http://localhost:8081"
	MAX_PEER           = 32

	mineAddress     string
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]*data.SignedTransaction)
	confirmedTx     = make(map[string]bool)
	SBC             data.SyncBlockChain
	Peers           data.PeerList
	ifStarted       = false
	heartBeatRevice = false
	keyString       = "00000"
	accounts        = make(map[string]*mpt.MerklePatriciaTrie)
	ContractAddress = "636f6e7472616374"
	MinerAddresses  = []string{
		"0480f8b87e2e2caedab0af1fe2975317de5c8d3146c515493ce21c9d90616923bfa12069a22afee549bb3b666af9dfe26e2543a2ba0263fecaaa60e38cdb0221d1",
		"041f5c565ffee4cf280295d864efd438e704894179cfdc552d83961f6eda877ad55cb2c46048c6cb96749a9b27c5ccdd7717b8647cccda07f145705879585ac271",
	}
	UserAddresses = []string{
		"0461ca35768e3960c76e881c2b2f543ae6c298f78fba9473f5b0da57f735dbb317f993575602ff7b6969fa423fb12b8fad88a9c1575aa5bc479d47ce0287ec7b2f",
		"04b31f6e431138d297df326efe3276e1dc7fd5aba95a8a303a8ebcbf81614dfba409f3da10b4a3537e8cc5ad30cbace338c25104a11a90f80820a4f79707ebbc58",
	}
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
	currentAccounts := mpt.MerklePatriciaTrie{}
	currentAccounts.Initial()

	gameContract := bc.Account{1000, ""}
	gameContractJSON, _ := gameContract.EncodeToJSON()
	currentAccounts.Insert(ContractAddress, gameContractJSON)

	for _, address := range MinerAddresses {
		mineAccount := bc.Account{1000, ""}
		mineAccountJSON, _ := mineAccount.EncodeToJSON()
		currentAccounts.Insert(address, mineAccountJSON)
	}
	for _, address := range UserAddresses {
		userAccount := bc.Account{1000, ""}
		userAccountJSON, _ := userAccount.EncodeToJSON()
		currentAccounts.Insert(address, userAccountJSON)
	}
	accounts[currentAccounts.Root] = &currentAccounts
	createGenesisBlock(currentAccounts.Root)
}

//Start will Register ID, download BlockChain, start HeartBeat
func StartMiner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["publicKey"]
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
	mineAddress = address
	//Start sending HeartBeat
	go startHeartBeat()
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

// Upload blockchain to whoever called this method, return jsonStr
func UploadAccounts(w http.ResponseWriter, r *http.Request) {
	var result []string
	fmt.Println("Num of accounts trie:", len(accounts))
	for root, accounts := range accounts {
		buffer := accounts.Serialize()
		fmt.Println("Root:", root)
		result = append(result, hex.EncodeToString(buffer))
	}
	resultBytes, _ := json.Marshal(result)
	fmt.Fprint(w, string(resultBytes))
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
		block, _ := bc.DecodeFromJSON(heartBeat.BlockJson)
		//Check PoW
		y := proofOfWork(block.Header.ParentHash, block.Header.Nonce, block.Value.Root)
		if y[:len(keyString)] != keyString {
			return
		}
		fmt.Println("VALID BLOCK")
		if block.Header.ParentHash != "" && !SBC.CheckParentHash(block) {
			askForBlock(block.Header.Height-1, block.Header.ParentHash)
		}
		parentBlock, err := SBC.GetParentBlock(block)
		if !err {
			fmt.Println("PARENT BLOCK NOT FOUND")
			return
		}
		valid, txsID := validateReceiveTxsInBlock(block.Value, parentBlock.Header.AccountsRoot)
		if valid {
			SBC.Insert(block)
			newAccountTrie := data.AddTransaction(accounts[parentBlock.Header.AccountsRoot].Copy(), 
				block.Value)
			//Add miner reward
			minerAccountJSON, _ := newAccountTrie.Get(block.Header.MinedBy)
			minerAccount, _ := bc.DecodeAccountFromJSON(minerAccountJSON)
			minerAccount.Balance = minerAccount.Balance + block.Header.BlockReward
			minerAccountJSON, _ = minerAccount.EncodeToJSON()
			newAccountTrie.Insert(block.Header.MinedBy, minerAccountJSON)
			if newAccountTrie.Root != block.Header.AccountsRoot {
				fmt.Println("WEIRD")
			} else {
				accounts[newAccountTrie.Root] = &newAccountTrie
			}
			deleteConfirmedTx(txsID)
			heartBeatRevice = true
		} else {
			fmt.Println("Txs IN BLOCK INVALID")
		}
	}
	if heartBeat.Hops > 1 {
		heartBeat.Hops--
		forwardHeartBeat(heartBeat)
	}
}

//Canonical will prints the current canonical chain, and chains of all forks if there are forks
func Canonical(w http.ResponseWriter, r *http.Request) {
	rs := ""
	blocks := SBC.GetLatestBlocks()
	for index, block := range blocks {
		currBlock := block
		rs += fmt.Sprintf("Chain #%d\n", index+1)
		for currBlock.Header.Hash != "" {
			rs += fmt.Sprintf("height=%d, hash=%s, parentHash=%s, accountsRoot=%s\n",
				currBlock.Header.Height, currBlock.Header.Hash,
				currBlock.Header.ParentHash, currBlock.Header.AccountsRoot)
			currBlock, _ = SBC.GetBlock(currBlock.Header.Height-1, currBlock.Header.ParentHash)
		}
		rs += "\n"
	}
	fmt.Fprint(w, rs)
}

func HandleTx(w http.ResponseWriter, r *http.Request) {
	buffer, _ := ioutil.ReadAll(r.Body)
	signedTransaction, _ := data.DecodeSignedTransactionFromJSON(string(buffer))
	validSign, _ := wallet.ValidateTxSignature(signedTransaction.Transaction, signedTransaction.Signature)
	if !validSign {
		fmt.Println("Invalid signature")
		return
	}
	//Received tx is already confirmed
	fmt.Println("Received: ", signedTransaction.Transaction.ID)
	if confirmedTx[signedTransaction.Transaction.ID] == true {
		fmt.Println("Tx confirmed")
		return
	}
	memoryPool[signedTransaction.Transaction.ID] = &signedTransaction
	startTryingNonces()
	// if len(memoryPool) >= 2 && len(mineAddress) > 0 {
	// 	startTryingNonces()
	// }
}

func initPeer(portNumber int32) {
	//Register
	RegisterLocal(portNumber)
	SBC = data.NewBlockChain()
}

func RegisterLocal(id int32) {
	Peers = data.NewPeerList(id, int32(MAX_PEER))
}

func Download() {
	DownloadPeerList()
	DownloadBC()
	DownloadAccounts()
}

func DownloadPeerList() {
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
}

func DownloadBC() {
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

func DownloadAccounts() {
	for addr := range Peers.Copy() {
		response, err := http.Get(addr + "/accounts/upload")
		if err == nil {
			var listAccountTrie []string
			buffer, _ := ioutil.ReadAll(response.Body)
			json.Unmarshal(buffer, &listAccountTrie)
			for _, accountTrieStr := range listAccountTrie {
				accountsByte, _ := hex.DecodeString(accountTrieStr)
				mpt := mpt.DeserializeMPT(accountsByte)
				accounts[mpt.Root] = &mpt
			}
			break
		} else {
			panic(err)
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
			parentBlock, _ := bc.DecodeFromJSON(string(blockBuffer))
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
		heartBeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), bc.Block{}, peersJSON, SELF_ADDR)
		heartBeatJSON, _ := heartBeat.EncodeToJSON()
		sendHeartBeat(heartBeatJSON)
	}
}

func sendHeartBeat(heartBeatJSON string) {
	peers := Peers.Copy()
	for addr := range peers {
		currAddr := addr
		go func() {
			_, err := http.Post(currAddr+"/heartbeat/receive", "application/json", bytes.NewBuffer([]byte(heartBeatJSON)))
			if err != nil {
				fmt.Println("Peer not available")
				Peers.Delete(currAddr)
			}
		}()
	}
}

//startTryingNonces tell miner to mine
func startTryingNonces() {
	//Solve puzzle
	for len(memoryPool) > 0 {
		rand.Seed(time.Now().UnixNano())
		y := ""
		//Get parents block
		parentBlock := SBC.GetLatestBlocks()[0]
		//Create MPT
		txMPT, txIDs := createMPTFromMemPool(parentBlock.Header.AccountsRoot)
		//TODO: Add transaction fee for the miner

		parentHash := parentBlock.Header.Hash
		//Generate nonce
		random, _ := utils.RandomHex(16)
		x := hex.EncodeToString(random)
		y = proofOfWork(parentHash, x, txMPT.Root)
		//Verify nonce
		for y[:len(keyString)] != keyString && !heartBeatRevice {
			random, _ = utils.RandomHex(16)
			x = hex.EncodeToString(random)
			y = proofOfWork(parentHash, x, txMPT.Root)
		}
		if heartBeatRevice {
			heartBeatRevice = false
			fmt.Println("Receive Block")
			continue
		} else {
			fmt.Print("Found Block: ")
			Peers.Rebalance()
			//Update balance in accounts Trie, delete the old root in accounts hashmap
			newAccountTrie := data.AddTransaction(accounts[parentBlock.Header.AccountsRoot].Copy(), txMPT)
			newBlock := SBC.GenBlock(txMPT, x, parentHash, mineAddress, &newAccountTrie)
			accounts[newAccountTrie.Root] = &newAccountTrie
			peersJSON, _ := Peers.PeerMapToJson()
			heartBeat := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), newBlock, peersJSON, SELF_ADDR)
			heartBeatJSON, _ := heartBeat.EncodeToJSON()
			sendHeartBeat(heartBeatJSON)
			deleteConfirmedTx(txIDs)
		}
	}
}

func createMPTFromMemPool(prevBalanceRoot string) (mpt.MerklePatriciaTrie, []string) {
	result := mpt.MerklePatriciaTrie{}
	var transactionsID []string
	result.Initial()
	//Get  transaction from memoryPool, check if it not confirmed yet
	//Check address to see if it is gameAddress or not
	for id, signedTransaction := range memoryPool {
		if confirmedTx[id] == true {
			delete(memoryPool, id)
			continue
		}
		if validateTx(signedTransaction, prevBalanceRoot) {
			signedTxJSON, _ := signedTransaction.EncodeToJSON()
			result.Insert(id, signedTxJSON)
		}
		transactionsID = append(transactionsID, id)
	}
	return result, transactionsID
}

func validateReceiveTxsInBlock(txsTrie mpt.MerklePatriciaTrie, prevBalanceRoot string) (bool, []string) {
	signedTxs := data.GetSignedTxsFromMPT(txsTrie)
	var txsID []string
	fmt.Println(prevBalanceRoot)
	for _, signedTx := range signedTxs {
		if confirmedTx[signedTx.Transaction.ID] {
			continue
		}
		if !validateTx(&signedTx, prevBalanceRoot) {
			return false, nil
		}
		txsID = append(txsID, signedTx.Transaction.ID)
	}
	return true, txsID
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

func createGenesisBlock(accountRoot string) {
	mpt := mpt.MerklePatriciaTrie{}
	mpt.Initial()
	SBC.Insert(bc.Genesis(accountRoot))
}

func validateBalance(publicKey string, amount int, prevBalanceRoot string) bool {
	accountJSON, _ := accounts[prevBalanceRoot].Get(publicKey)
	if accountJSON == "" {
		return false
	}
	account, _ := bc.DecodeAccountFromJSON(accountJSON)
	if account.Balance <= amount {
		return false
	}
	return true
}

func validateTx(signedTx *data.SignedTransaction, prevBalanceRoot string) bool {
	result, _ := wallet.ValidateTxSignature(signedTx.Transaction, signedTx.Signature)
	if !result {
		fmt.Println("Signature not correct")
		return false
	}
	from := string(signedTx.Transaction.FromAddress)
	to := string(signedTx.Transaction.FromAddress)
	// fmt.Println(accounts[prevBalanceRoot].String())
	fromAccountJSON, _ := accounts[prevBalanceRoot].Get(from)
	if fromAccountJSON == "" {
		fmt.Println("fromAccountJSON is empty")
		return false
	}
	fromAccount, _ := bc.DecodeAccountFromJSON(fromAccountJSON)
	if fromAccount.Balance <= signedTx.Transaction.Value {
		fmt.Println("balance not enough")
		return false
	}
	toAccountJSON, _ := accounts[prevBalanceRoot].Get(to)
	//If toAccount is empty, create account with balance = 0
	if toAccountJSON == "" {
		fmt.Println("toAccountJSON is empty")
		return false
	}
	return true
}

func GetAccountFork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	publicKey := vars["publicKey"]
	blocks := SBC.GetLatestBlocks()
	for _, block := range blocks {
		account, valid := GetAccountByPublicKey(publicKey, block.Header.AccountsRoot)
		if valid{
			accountJSON,_:=account.EncodeToJSON()
			fmt.Fprint(w, accountJSON)
		}
	}
}

func ShowBalance(w http.ResponseWriter, r *http.Request) {
	blocks := SBC.GetLatestBlocks()
	for _, block := range blocks {
		fmt.Fprintf(w, "%s\n", accounts[block.Header.AccountsRoot].String())
	}
}

func GetAccountByPublicKey(publicKey string, accountsRoot string) (bc.Account, bool) {
	accountsTrie := accounts[accountsRoot]
	// for key, value := range accountsTrie.KeyVal {
	// 	fmt.Println(key, " - ", value)
	// }
	accountJSON, _ := accountsTrie.Get(publicKey)
	account, err := bc.DecodeAccountFromJSON(accountJSON)
	if accountJSON == "" || err != nil {
		return account, false
	}
	return account, true
}

func GetGameInformation(w http.ResponseWriter, r *http.Request) {
	blocks := SBC.GetLatestBlocks()
	for _, block := range blocks {
		accountsTrie := accounts[block.Header.AccountsRoot]
		accountJSON, _ := accountsTrie.Get(ContractAddress)
		account, _ := bc.DecodeAccountFromJSON(accountJSON)
		gameList, _ := data.DecodeGameListFromJSON(account.Data)
		fmt.Fprintf(w, "Balance: %d\n", account.Balance)
		for _, game := range gameList {
			gameJSON, _ := game.EncodeToJSON()
			fmt.Fprintf(w, "%s\n", gameJSON)
		}
	}
}
