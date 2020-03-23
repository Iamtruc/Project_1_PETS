package main

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestEval(t *testing.T) {
	for i:= 0;i<len(TestCircuits);i++{
		fmt.Println("circuit"+strconv.Itoa(i), i)
		t.Run("circuit"+strconv.Itoa(i),func(t *testing.T) {
			my_circuit := TestCircuits[i]
			peers := my_circuit.Peers
			NbMult := 0 //Counts the number of multiplications in the circuit.
			N := uint64(len(peers))
			P := make([]*LocalParty, N, N)
			dummyProtocol := make([]*DummyProtocol, N, N)
			beaverprotocol := make([]*BeaverProtocol, N, N)

			for _, element := range my_circuit.Circuit {
				if element.Identify() == "Mult" {
					NbMult++
				}
			}
			var err error
			wg := new(sync.WaitGroup)
			for i := range peers {
				P[i], err = NewLocalParty(i, peers)
				P[i].WaitGroup = wg
				check(err)

				dummyProtocol[i] = P[i].NewDummyProtocol(uint64(i + 10))
				if NbMult > 0 {
					beaverprotocol[i] = P[i].NewBeaverProtocol()
					dummyProtocol[i].BeaverProt = beaverprotocol[i]
					beaverprotocol[i].ID = dummyProtocol[i].ID
					beaverprotocol[i].Generate_input(N)
				}
				}


			network := GetTestingTCPNetwork(P)

			for i, Pi := range dummyProtocol {
				Pi.BindNetwork(network[i])
			}
			if NbMult >0 {
				network = GetTestingTCPNetwork(P)
				for i, Pi := range beaverprotocol {
					Pi.BindNetwork(network[i])
				}
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

			correct:= 0
			for _, p := range dummyProtocol {
				if p.Output == my_circuit.ExpOutput{
					correct++
				}
			}

			switch correct {
			case int(N):
				fmt.Println("test completed")
			default:
				fmt.Println("Failed")
		}
	})
	}
}
