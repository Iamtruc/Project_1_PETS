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
	Beav.mod = 100// The mod is to be fixed by the peers. For the time being, it is constant
	var listA []uint64
	var listB []uint64
	var SumA uint64
	var SumB uint64
	var n = len(Beav.Peers)
	for i:=0;i<n;i++{
		A := uint64(rand.Intn(Beav.mod))
		B := uint64(rand.Intn(Beav.mod))
		listA = append(listA, A)
		listB = append(listB, B)
		SumA += A
		SumB += B
	}

	var C = SumA * SumB
	var little_c uint64
	var listC []uint64
	for i:=0;i<n;i++{
		little_c = uint64(rand.Intn(int(C)/(n-i)+1))
		listC = append(listC, little_c)
		C -=little_c
	}
	listC[0] += C

	for _, peer := range Beav.Peers{
		if peer.ID != Beav.ID{
			peer.BeaverChannel <- BeaverMessage{0, listA[int(peer.ID)]  }
			peer.BeaverChannel <- BeaverMessage{1, listB[int(peer.ID)] }
			peer.BeaverChannel <- BeaverMessage{2, listC[int(peer.ID)]  }
		}
	}

	Beav.MyInputs.ListA = listA[0]
	Beav.MyInputs.ListB = listB[0]
	Beav.MyInputs.ListC = listC[0]
}

func (Beav *BeaverProtocol) Run() (a,b,c uint64){

	if Beav.ID != 0{
		received := 0
		for m:= range Beav.BeaverChannel{
			switch m.index {
			case 0:
				Beav.MyInputs.ListA = m.value
			case 1:
				Beav.MyInputs.ListB = m.value
			case 2:
				Beav.MyInputs.ListC = m.value
			}
			received++
			if received == 3{
				break
			}
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