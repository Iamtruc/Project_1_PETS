package main

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

// go test -v -run=TestEval/circuit6

func TestEval(t *testing.T) {
	// We start by identifying what is the request of the user (which circuit he wants to test)
	for i:= 0;i<len(TestCircuits);i++{

		fmt.Println("circuit"+strconv.Itoa(i), i)

		// We then test the identified circuit
		t.Run("circuit"+strconv.Itoa(i),func(t *testing.T) {

			// We start by creating all the variables used in the circuit
			my_circuit := TestCircuits[i]
			peers := my_circuit.Peers
			NbMult := 0 //Counts the number of multiplications in the circuit.
			N := uint64(len(peers))
			P := make([]*LocalParty, N, N)
			dummyProtocol := make([]*DummyProtocol, N, N)
			beaverprotocol := make([]*BeaverProtocol, N, N)

			// We count the number of multiplications to know how many triplets we will need to create (unused variable for the trusted third party architecture)
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
					go beaverprotocol[i].Generate_input()
				}
			}
			time.Sleep(time.Second)

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
			time.Sleep(time.Second)

			// Once all the Dummyprotocols are bound together, we make them split their shares with the other peers and read the circuit.

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


			// Finally, we check if the outcome of the circuit is indeed the expected outcome.
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
