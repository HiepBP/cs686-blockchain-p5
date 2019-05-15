# cs686-blockchain-p4

  

<p>In this project, I built a web application base on blockchain that support user to play Rock Paper Scissors with each other. My Web application will work as a full-node in a network.</p>

  

![alt text](https://i.imgur.com/DwJa62g.jpg)

  

<p>The network contain a web server; a beginning full node that generate genesis block and every other nodes will come there to get peer list and blockchain when start; a list of miners to mine block and validate transaction.</p>

<p>Every node in the network will maintain a blockchain, a accounts information list, a peer slist, a memory pool for unvalidated transaction.</p>
  
  
# Important Class
## Transaction
| Variable 	| Type 	| Description 	|
|-------------	|--------	|----------------------------------------	|
| ID 	| string 	| ID is a hash of the transaction detail 	|
| FromAddress 	| string 	|  	|
| ToAddress 	| string 	|  	|
| Value 	| int 	| Value of the transaction 	|
| Data 	| string 	| The data of function in GameContract 	|
| Timestamp 	| int64 	|  	|
| Fee 	| int 	| Fee for transaction 	|

## SignedTransaction
| Variable 	| Type 	| Description 	|
|-------------	|--------	|----------------------------------------	|
| Transaction 	| Transaction 	|  	|
| Signature 	| []byte 	| The signature of current transaction 	|

## Block


| Block 	| Type 	| Description 	|
|--------------	|--------	|------------------------------------------------------	|
| Height 	| int32 	|  	|
| Timestampt 	| in64 	|  	|
| Hash 	| string 	|  	|
| ParentHash 	| string 	|  	|
| Size 	| int32 	|  	|
| Nonce 	| string 	|  	|
| AccountsRoot 	| string 	| MPT root of account trie 	|
| MinedBy 	| string 	| Address of miner who create a block 	|
| BlockReward 	| int 	| The value of block reward (default reward + txs fee) 	|

## Functions

### 1. Create Game

<p>User 1 will user Web UI to create a new game. He will select a choice, amount of coin, a secret number that will be used to hash choice and a private key. All of this information will be sign with private key in client side and send to web server. After that web server will send this data to other peer in peer list.</p>

### 2. Join Game

<p>User 2 will user Web UI to join a existing game. He will select a choice and input private key. All of this information will be sign with private key in client side and send to web server. After that web server will send this data to other peer in peer list.</p>

### 3. Reveal choice/close game

<p>User 1 will user Web UI to reveal his choice. He will select a choice and his secret number. All of this information will be sign with private key in client side and send to web server. After that web server will send this data to other peer in peer list.</p>

### 4. Mine transaction

<p>When number of unvalidated tx in memory pool larger than a specific number, miner will get tx from that, validate each tx and put the validated tx into mpt. Then he need to solve the PoW and put the tx list with nonce number into new block and forward that block to other node.</p>

  
  

![alt text](https://i.imgur.com/5ckTWIB.jpg)
How transaction be signed and send through network.<br/>

  

## Progress

[x] Basic UI

[x] Register and generate public, private key

[x] Sign Tx function

[x] Validate Tx is signed by a correct one

[x] Send account(public key) to other peers when create.

[x] See balance base on public key

[x] Classify Tx (Transfer coin Tx or Logic Tx)

[x] Miner validate transaction is signed correctly

[x] Miner validate balance for each transaction

[x] Miner handle logic code to create game

[x] Miner handle logic code to join game

[x] Miner handle logic code to reveal choice and close game

[x] User can see the game information

# Overall
The application still have some problem
1/ When Dealer create a game, if not one join it, dealer can not retrieve the money back
2/ When Player join a game, if Player join a game and after that, Dealer does not reveal his choice, both Dealer and Player will not receive money back
3/ The PoW algorithm need to be able to change the difficult base on the running time to generate previous block. Because i just set the constant difficult, so sometime it will too hard, sometime it will too easy.
