package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)
// The main file takes [Party ID] and [Inputs] as arguments and evaluates the circuit names NewMyCircuit
// The circuit is hard coded for the moment.
var NewMyCircuit = []Operation{
	&Input{
		Party: 0,
		Out:   0,
	},
	&Input{
		Party: 1,
		Out:   1,
	},
	&Input{
		Party: 2,
		Out:   2,
	},
	&Add{
		In1: 0,
		In2: 1,
		Out: 3,
	},
	&Add{
		In1:  2,
		In2:  3,
		Out : 4,
	},
	&MultCst{
		In: 4,
		CstValue: 5,
		Out:  5,
	},
	&Reveal{
		In:  5,
		Out: 6,
	},
}

func main() {
	prog := os.Args[0]
	args := os.Args[1:]

	if len(args) < 2 {
		fmt.Println("Usage:", prog, "[Party ID] [Input]")
		os.Exit(1)
	}

	partyID, errPartyID := strconv.ParseUint(args[0], 10, 64)
	if errPartyID != nil {
		fmt.Println("Party ID should be an unsigned integer")
		os.Exit(1)
	}

	partyInput, errPartyInput := strconv.ParseUint(args[1], 10, 64)
	if errPartyInput != nil {
		fmt.Println("Party input should be an unsigned integer")
		os.Exit(1)
	}

	Client(PartyID(partyID), partyInput)
}

func Client(partyID PartyID, partyInput uint64) {

	//N := uint64(len(peers))
	peers := map[PartyID]string {
		0: "localhost:6660",
		1: "localhost:6661",
		2: "localhost:6662",
	}

	// Create a local party 
	lp, err := NewLocalParty(partyID, peers)
	check(err)

	// Create the network for the circuit
	network, err := NewTCPNetwork(lp)
	check(err)

	// Connect the circuit network 
	err = network.Connect(lp)
	check(err)
	fmt.Println(lp, "connected")
	<- time.After(time.Second) // Leave time for others to connect

	// Create a new circuit evaluation protocol 
	dummyProtocol := lp.NewDummyProtocol(partyInput)
	// Bind evaluation protocol to the network
	dummyProtocol.BindNetwork(network)

	// Creating the beaverprotocol
	beaverprotocol := lp.NewBeaverProtocol()
	dummyProtocol.BeaverProt = beaverprotocol
	beaverprotocol.ID = dummyProtocol.ID

	// Create the network for the beaverprotocols
	network2, err := NewTCPNetwork(lp)
	check(err)

	// Connect the beaverprotocol network
	err = network2.Connect(lp)
	check(err)
	beaverprotocol.BindNetwork(network2)

	// We now have to split our share among our participants.
	dummyProtocol.peerInput = make(map[PartyID]uint64)
	dummyProtocol.peerCircuit = make(map[WireID]uint64)

	// Split the share between the different peers
	var truc = map[PartyID]map[GateID]uint64{partyID:{GateID(partyID):partyInput}}
	dummyProtocol.Splitshare(truc)
	time.Sleep(time.Second/4)

	// Evaluate the circuit
	time.Sleep(time.Second/4)
	dummyProtocol.readcircuit(NewMyCircuit)

	fmt.Println(lp, "completed with output", dummyProtocol.Output)
	}
