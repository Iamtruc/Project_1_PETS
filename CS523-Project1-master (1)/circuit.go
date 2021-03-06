package main

import (
	"fmt"
	"math/rand"
	"time"
)

func (cep *DummyProtocol) Splitshare(Inputs map[PartyID]map[GateID]uint64)(){

	// We start by getting the element in Input associated to cep's ID
	for _,element := range Inputs[cep.ID] {
		cep.peerInput[cep.ID] = element
	}

	// We then generate our shares.
	my_len := len(cep.Peers)
	var list_split []uint64
	var leftover uint64 = cep.peerInput[cep.ID]
	for i:=0;i<my_len-1;i++{
		if (int(cep.peerInput[cep.ID])/my_len) != 0{
			list_split = append(list_split, uint64(rand.Intn( int(cep.peerInput[cep.ID])/my_len)))
		}
		if (int(cep.peerInput[cep.ID])/my_len) == 0{
			list_split = append(list_split, 0)
		}
		leftover -= list_split[i]
	}
	cep.peerInput[cep.ID] = leftover

	// Once we have generated our shares, we can broadcast them to the other peers
	i:=0
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan <- DummyMessage{cep.ID, list_split[i]}
			i++
		}
	}

	// We then wait for them to send their shares back.
	received := 0
	for m := range cep.Chan {
		cep.peerInput[m.Party] = m.Value
		received++
		if received == len(cep.Peers)-1 {
			break
		}
	}

	if cep.WaitGroup != nil {
		cep.WaitGroup.Done()
	}
}

func (cep *DummyProtocol) readcircuit(circuit []Operation){
	// We take a circuit as input and read the circuit until the end.
	for _,op := range circuit{
		err, name := op.canEval(*cep)
		// We check for errors, and then, we check for the name of the gate.
		if (err == nil){
			switch name{
			// If we have a multiplication gate, we get a Beaver triplet
			case "Mult":
				switch cep.ID{
				case 0:
					//cep.BeaverProt.GenInput()
					//cep.BeaverA, cep.BeaverB, cep.BeaverC = cep.BeaverProt.Run()
					time.Sleep(time.Second/5)
					op.Eval(*cep)
				default:
					//cep.BeaverA, cep.BeaverB, cep.BeaverC = cep.BeaverProt.Run()
					time.Sleep(time.Second/5)
					op.Eval(*cep)
				}
				// If we have a reveal gate, we update our output
			case "Reveal":
				cep.Output = op.Eval(*cep)
				if cep.WaitGroup != nil {
					cep.WaitGroup.Done()
				}
			// In the general case, we just run the gate
			default:
				op.Eval(*cep)
			}
		}
		// If we have an error, print a subtle message to inform user.
		if (err != nil){
			fmt.Println("Cataschtroumpf")
		}
	}
}