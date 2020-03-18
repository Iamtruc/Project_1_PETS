package main

import (
	"math/rand"
)

type BeaverProtocol struct {
	*LocalParty
	Chan chan BeaverMessage
	Peers map[PartyID]*DummyRemote
	PeerA map [PartyID][]uint64
	PeerB map [PartyID][]uint64
	PeerC map [PartyID][]uint64
}

type BeaverRemoteParty struct{
	*RemoteParty
	Chan chan BeaverMessage
}

type BeaverMessage struct{
	a uint64
	b uint64
	c uint64
}

type BeaverInputs struct{

}

func (lp *LocalParty) NewBeaverProtocol(input uint64) *DummyProtocol {
	beav := new(BeaverProtocol)
	beav.LocalParty = lp
	beav.Chan = make(chan BeaverMessage, 32)
	beav.Peers = make(map[PartyID]*DummyRemote, len(lp.Peers))
	for i, rp := range lp.Peers {
		cep.Peers[i] = &DummyRemote{
			RemoteParty:  rp,
			Chan:         make(chan DummyMessage, 32),
		}
	}

	return cep
}


func generate_beaver(n,mod int) ([]uint64, []uint64, uint64){
	var list_a []uint64
	var list_b []uint64
	var a uint64
	var b uint64
	var my_sum uint64
	my_sum = 1
	for i:=0; i<n;i++{
		a = uint64(rand.Intn(mod))
		b = uint64(rand.Intn(mod))
		list_a = append(list_a, a)
		list_b = append(list_b, b)
		my_sum += b
	}
	return list_a, list_b, my_sum
}

func (cep *DummyProtocol)ask_beaver() {
	mod := 5
	n := len(cep.Peers)
	list_a, list_b, my_sum := generate_beaver(n, mod)
	for _, peer := range cep.Peers {
			peer.Chan<- DummyMessage{PartyID(n + 3), list_a[int(peer.ID)]}
			peer.Chan<- DummyMessage{PartyID(n + 4), list_b[int(peer.ID)]}
			peer.Chan<- DummyMessage{PartyID(n + 5), list_a[int(peer.ID)] * my_sum}
		}
}