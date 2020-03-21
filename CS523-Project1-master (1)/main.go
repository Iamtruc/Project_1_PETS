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
	&Mult{
		In1: 0,
		In2: 1,
		Out: 3,
	},
	&Mult{
		In1:  0,
		In2:  1,
		Out : 4,
	},
	&Mult{
		In1:  0,
		In2:  2,
		Out : 5,
	},
	&Add{
		In1: 3,
		In2: 4,
		Out: 6,
	},
	&Add{
		In1: 5,
		In2: 6,
		Out: 7,
	},
	&Reveal{
		In:  7,
		Out: 8,
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

	Client(PartyID(partyID), partyInput, NewMyCircuit)
}



// Unused

func Client(partyID PartyID, partyInput uint64, evalCircuit []Operation) uint64{

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

	//dummyProtocol.WaitGroup = wg

	// Creating the beaverprotocol
	beaverprotocol := lp.NewBeaverProtocol()
	dummyProtocol.BeaverProt = beaverprotocol
	beaverprotocol.ID = dummyProtocol.ID

	// We now have to split our share among our participants.
	dummyProtocol.peerInput = make(map[PartyID]uint64)
	dummyProtocol.peerCircuit = make(map[WireID]uint64)
	var truc = map[PartyID]map[GateID]uint64{partyID:{GateID(partyID):partyInput}}
	dummyProtocol.Splitshare(truc)
	<- time.After(time.Second)

	// Evaluate the circuit
	dummyProtocol.readcircuit(evalCircuit)
	<- time.After(time.Second)

	return(dummyProtocol.Output)
}
