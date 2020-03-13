package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestEval(t *testing.T) {
	for _,my_circuit := range TestCircuits{
	peers := my_circuit.Peers
	N := uint64(len(peers))
	P := make([]*LocalParty, N, N)
	dummyProtocol := make([]*DummyProtocol, N, N)

	var err error
	wg := new(sync.WaitGroup)
	for i := range peers {
		P[i], err = NewLocalParty(i, peers)
		P[i].WaitGroup = wg
		check(err)

		dummyProtocol[i] = P[i].NewDummyProtocol(uint64(i + 10))
	}

	network := GetTestingTCPNetwork(P)
	fmt.Println("parties connected")

	for i, Pi := range dummyProtocol {
		Pi.BindNetwork(network[i])
	}

	for _, p := range dummyProtocol {
		p.peerInput = make(map[PartyID]uint64)
		p.peerCircuit = make(map[WireID]uint64)
		p.Add(1)
		go p.Splitshare(my_circuit.Inputs)
	}
	wg.Wait()
		for _, p := range dummyProtocol {
			p.Add(1)
			go p.readcircuit(my_circuit.Circuit)
		}
		wg.Wait()

	fmt.Println("test completed")
}
}
