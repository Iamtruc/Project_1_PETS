package main

import (
	"encoding/binary"
	"math/rand"
	"net"
)

type BeaverProtocol struct {
	*LocalParty
	BeaverChannel chan BeaverMessage
	Peers map[PartyID]*BeaverRemoteParty
	MyInputs BeaverInputs
	mod int
	ID PartyID
}

type BeaverRemoteParty struct{
	*RemoteParty
	BeaverChannel chan BeaverMessage
}

type BeaverMessage struct{
	index uint64
	value uint64
}

type BeaverInputs struct{
	ListA uint64
	ListB uint64
	ListC uint64

}

func (lp *LocalParty) NewBeaverProtocol() *BeaverProtocol {
	beav := new(BeaverProtocol)
	beav.LocalParty = lp
	beav.BeaverChannel = make(chan BeaverMessage, 32)
	beav.Peers = make(map[PartyID]*BeaverRemoteParty, len(lp.Peers))
	for i, rp := range lp.Peers {
		beav.Peers[i] = &BeaverRemoteParty{
			RemoteParty:  rp,
			BeaverChannel:         make(chan BeaverMessage, 32),
		}
	}
	return beav
}

func (Beav *BeaverProtocol)GenInput(){
	Beav.mod = 100
	A := uint64(rand.Intn(Beav.mod))// The mod is to be fixed by the peers. For the time being, it is constant
	B := uint64(rand.Intn(Beav.mod))
	Beav.MyInputs.ListA = A
	Beav.MyInputs.ListB = B
	Beav.MyInputs.ListC = B
}

func (Beav *BeaverProtocol) Run() (a,b,c uint64){
	Beav.GenInput()

	for _, peer := range Beav.Peers{
		if peer.ID != Beav.ID{
			peer.BeaverChannel <- BeaverMessage{2, Beav.MyInputs.ListC  }
		}
	}

	received := 0
	for m:= range Beav.BeaverChannel{
		if m.index == 2{
			Beav.MyInputs.ListC += m.value
		}
		received++
		if received == len(Beav.Peers)-1{
			break
		}
	}

	return Beav.MyInputs.ListA, Beav.MyInputs.ListB, Beav.MyInputs.ListC
}

func (Beav *BeaverProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		rp := Beav.Peers[partyID]

		// Receiving loop from remote
		go func(conn net.Conn, rp *BeaverRemoteParty) {
			for {
				var val uint64
				var ind uint64
				var err error
				err = binary.Read(conn, binary.BigEndian, &ind)
				check(err)
				err = binary.Read(conn, binary.BigEndian, &val)
				check(err)
				msg := BeaverMessage{
					index: ind,
					value: val,
				}
				//fmt.Println(cep, "receiving", msg, "from", rp)
				Beav.BeaverChannel <- msg
			}
		}(conn, rp)

		// Sending loop of remote
		go func(conn net.Conn, rp *BeaverRemoteParty) {
			var m BeaverMessage
			var open = true
			for open {
				m, open = <- rp.BeaverChannel
				//fmt.Println(cep, "sending", m, "to", rp)
				check(binary.Write(conn, binary.BigEndian, m.index))
				check(binary.Write(conn, binary.BigEndian, m.value))
			}

		}(conn, rp)
	}
}

func (Beav *BeaverProtocol)exchangexminusa(xminusa, yminusb uint64) (uint64, uint64){
	for _, peer := range Beav.Peers{
		if peer.ID != Beav.ID{
			peer.BeaverChannel <- BeaverMessage{0, xminusa }
			peer.BeaverChannel <- BeaverMessage{1, yminusb }
		}
	}

	received := 0
	for m:= range Beav.BeaverChannel{
		if m.index == 0{
			xminusa += m.value
		}
		if m.index == 1{
			yminusb += m.value
		}
		received++
		if received == 2 * len(Beav.Peers)-2{
			break
		}
	}

	return xminusa, yminusb
}