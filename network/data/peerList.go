package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

//PeerList contains information of current peer and others
type PeerList struct {
	SelfId    int32            `json:"selfId"`
	PeerMap   map[string]int32 `json:"peerMap"`
	MaxLength int32            `json:"maxLength"`
	mux       sync.Mutex
}

//NewPeerList create a peer base on id and maxLength
func NewPeerList(id int32, maxLength int32) PeerList {
	peerList := PeerList{PeerMap: make(map[string]int32), MaxLength: maxLength}
	peerList.Register(id)
	return peerList
}

//Add help to add new peer(addr && id) to current peer
func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	peers.PeerMap[addr] = id
	peers.mux.Unlock()
}

//Delete specific peer from current peer
func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.PeerMap, addr)
	peers.mux.Unlock()
}

//Rebalance will keep maxLength closet peers
//Sort all peers' Id, insert SelfId, consider the list as a cycle, and choose maxLength/2 nodes at each side of SelfId
func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	if len(peers.PeerMap) <= int(peers.MaxLength) {
		peers.mux.Unlock()
		return
	}
	numPeer := len(peers.PeerMap)
	peerMapByID := map[int32]string{}
	var listID []int32
	for k, v := range peers.PeerMap {
		peerMapByID[v] = k
		listID = append(listID, v)
	}
	sort.Slice(listID, func(i, j int) bool { return listID[i] < listID[j] })
	left := 0
	right := 0
	for index, ID := range listID {
		if ID > peers.SelfId {
			left = ((index-int(peers.MaxLength)/2)%numPeer + numPeer) % numPeer
			right = (index - 1 + int(peers.MaxLength)/2) % numPeer
			fmt.Printf("Index: %d  -  Max Length: %d  -  numPeer: %d\n", index, peers.MaxLength, numPeer)
			fmt.Printf("Left: %d  -  Right: %d\n", left, right)
			break
		}
	}
	//Get the delete IDs
	for index, ID := range listID {
		if left < right {
			if index < left || index > right {
				addr := peerMapByID[ID]
				delete(peers.PeerMap, addr)
			}
		} else {
			if index < left && index > right {
				addr := peerMapByID[ID]
				delete(peers.PeerMap, addr)
			}
		}
	}
	peers.mux.Unlock()
}

//Show will return a string of PeerMap
func (peers *PeerList) Show() string {
	peers.mux.Lock()
	rs := ""
	for addr, id := range peers.Copy() {
		rs += fmt.Sprintf(" - addr=%s, id=%d\n", addr, id)
	}
	peers.mux.Unlock()
	rs = "This is PeerMap:\n" + rs
	return rs
}

//Register will add an Id to current PeerList
func (peers *PeerList) Register(id int32) {
	peers.mux.Lock()
	peers.SelfId = id
	peers.mux.Unlock()
	fmt.Printf("SelfId=%v\n", id)
}

//Copy will make a copy of PeerMap
func (peers *PeerList) Copy() map[string]int32 {
	copyMap := make(map[string]int32)
	for key, value := range peers.PeerMap {
		copyMap[key] = value
	}
	return copyMap
}

//GetSelfId will return PeerList id
func (peers *PeerList) GetSelfId() int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.SelfId
}

//PeerMapToJson will encode peerMap to JSON
func (peers *PeerList) PeerMapToJson() (string, error) {
	var result string
	peers.mux.Lock()
	peerListJSON, err := json.Marshal(peers.Copy())
	defer peers.mux.Unlock()
	if err != nil {
		return result, err
	}
	result = string(peerListJSON)
	return result, nil
}

//InjectPeerMapJson will copy received peerMap to current peerList
func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	peers.mux.Lock()
	var peerMap map[string]int32

	json.Unmarshal([]byte(peerMapJsonStr), &peerMap)
	for addr, id := range peerMap {
		if addr != selfAddr {
			peers.PeerMap[addr] = id
		}
	}
	peers.mux.Unlock()
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	// peerMapJSON, _ := peers.PeerMapToJson()
	// fmt.Println("JSON", peerMapJSON)
	// peers.InjectPeerMapJson("{\"2222\":10,\"3333\":13,\"5555\":12,\"6666\":212,\"4444\":7}", "2345")
	// fmt.Println(peers.Show())
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
