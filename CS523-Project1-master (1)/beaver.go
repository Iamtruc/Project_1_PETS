package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ldsec/lattigo/bfv"
	"github.com/ldsec/lattigo/ring"
	"math"
	"time"

	//"math/rand"
	"net"
)

var params = bfv.DefaultParams[bfv.PN13QP218]
var N = params.LogN
var T = params.T
var sigm = params.Sigma
//var Q = params.LogQi

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
	index PartyID
	value uint64
	my_byte []byte
}

type BeaverInputs struct{
	ListA []uint64
	ListB []uint64
	ListC []uint64

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

func (Beav *BeaverProtocol)Generate_input(nbBeaver uint64){
	var vecta = newRandomVec(N,T)
	var vectb = newRandomVec(N,T)
	var aPlaintext  = bfv.NewPlaintext(params)
	var bPlaintext  = bfv.NewPlaintext(params)
	var vectc = mulVec(vecta, vectb)
	var vectd = make([]*bfv.Ciphertext, N)
	var vectr = make([]uint64, N)
	var vectrenc = bfv.NewPlaintext(params)

	var Bound = uint64(math.Floor(6*sigm))
	var evaluator = bfv.NewEvaluator(params)
	var encoder = bfv.NewEncoder(params)
	kgen := bfv.NewKeyGenerator(params)
	var inter = bfv.NewCiphertext(params, N)// Used for intermediate calculations, no real meaning
	var inter2 = bfv.NewCiphertext(params, N)// Used for intermediate calculations, no real meaning
	var mess BeaverMessage
	var TravelPoly = make([]byte, 2)// This variable is called TravelPoly, because it's going to travel
	//through the BeaverChannels.
	var writeme = new(bytes.Buffer) // Buffer where we write data to then send them to the peers
	var BeaverCipherText = bfv.NewCiphertext(params, N)
	var to_send []BeaverMessage
	var index_send []PartyID
	var MyContext, err = ring.NewContextWithParams(uint64(1 << params.LogN),[]uint64{params.T, params.T, params.T})
	if err != nil {
		panic(err)
	}

	vectsk :=  kgen.GenSecretKey()
	encryptorBeaverSk := bfv.NewEncryptorFromSk(params, vectsk)
	decryptorBeaverSK := bfv.NewDecryptor(params, vectsk)

	encoder.EncodeUint(vecta, aPlaintext)
	encoder.EncodeUint(vectb, bPlaintext)

	BeaverCipherText = encryptorBeaverSk.EncryptNew(aPlaintext)
	vectd[int(Beav.ID)] = BeaverCipherText

	TravelPoly, err = BeaverCipherText.MarshalBinary()
	if err != nil {
		fmt.Println("Marshal failed:", err)
	}

	err = binary.Write(writeme, binary.BigEndian, TravelPoly)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	var buf = make([]byte, len(TravelPoly))
	copy(buf,writeme.Bytes())

	for _, peer := range Beav.Peers{
		if peer.ID != Beav.ID{
			mess = BeaverMessage{Beav.ID, 0, buf}
			peer.BeaverChannel <- mess
			time.Sleep(time.Second/2)
		}
	}

	var e = make([]*ring.Poly, 2)
	e[0] = ring.NewPoly(N, params.T)
	e[1] = ring.NewPoly(N, params.T)
	var ebfv = bfv.NewPlaintext(params)


	received := 0
	for m:= range Beav.BeaverChannel{

		fmt.Println(m)

		// We get a new Gaussian noise.
		e[0] = &ring.Poly{MyContext.SampleGaussianNew(sigm,Bound).GetCoefficients()}
		e[1] = &ring.Poly{MyContext.SampleGaussianNew(sigm,Bound).GetCoefficients()}
		ebfv.SetValue(e)

		vectd[int(m.index)] = bfv.NewCiphertext(params, 1)
		err = vectd[int(m.index)].UnmarshalBinary(UIntToByteSlice(m.value))
		if err != nil {
			fmt.Println("The Marshal failed:", err)
		}

		vectr = newRandomVec(N,T)
		vectc = subVec(vectc, vectr)
		encoder.EncodeUint(vectr, vectrenc)

		inter = evaluator.AddNew(evaluator.MulNew(vectd[int(m.index)], bPlaintext), vectrenc)
		inter = evaluator.AddNew(inter, ebfv)

		TravelPoly, err = inter.MarshalBinary()

		writeme.Reset()
		err = binary.Write(writeme, binary.BigEndian, TravelPoly)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}

		mess = BeaverMessage{Beav.ID, 0,buf}
		to_send = append(to_send, mess)
		index_send = append(index_send, m.index)
		received++
		if received == len(Beav.Peers)-1{
			break
		}
	}

	fmt.Println(to_send)

	for i, message := range to_send{
		time.Sleep(time.Second/2)
		Beav.Peers[index_send[i]].BeaverChannel <- message
	}

	received = 0

	var my_c = bfv.NewCiphertext(params, 1)
	for m:= range Beav.BeaverChannel {

		vectd[int(m.index)] = bfv.NewCiphertext(params, 1)
		err = inter2.UnmarshalBinary(UIntToByteSlice(m.value))
		if err != nil {
			fmt.Println("The Marshal failed:", err)
		}

		evaluator.Add(my_c, inter2, my_c)

		received++
		if received == len(Beav.Peers)-1 {
			break
		}
	}

	vectc = addVec(vectc, encoder.DecodeUint(decryptorBeaverSK.DecryptNew(my_c)))

	fmt.Println(vecta, Beav.ID)
	fmt.Println(vectb, Beav.ID)
	fmt.Println(vectc, Beav.ID)

	Beav.MyInputs.ListC = vectc
	Beav.MyInputs.ListA = vecta
	Beav.MyInputs.ListB = vectb

	if Beav.WaitGroup != nil {
		Beav.WaitGroup.Done()
	}
}

func (Beav *BeaverProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		rp := Beav.Peers[partyID]

		// Receiving loop from remote
		go func(conn net.Conn, rp *BeaverRemoteParty) {
			for {
				var val uint64
				var ind PartyID

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
				//check(binary.Write(conn, binary.BigEndian, m.ciphertext))
			}

		}(conn, rp)
	}
}
/*
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
			peer.BeaverChannel <- BeaverMessage{0, listA[int(peer.ID)] ,0 }
			peer.BeaverChannel <- BeaverMessage{1, listB[int(peer.ID)],0 }
			peer.BeaverChannel <- BeaverMessage{2, listC[int(peer.ID)],0  }
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
*/

func (Beav *BeaverProtocol)exchangexminusa(xminusa, yminusb uint64) (uint64, uint64){
	var truc []byte
	for _, peer := range Beav.Peers{
		if peer.ID != Beav.ID{
			peer.BeaverChannel <- BeaverMessage{0, xminusa , truc}
			peer.BeaverChannel <- BeaverMessage{1, yminusb , truc}
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


func UIntToByteSlice(nb uint64) []byte{
	buf := make([]byte, 393222)
	_ = binary.PutUvarint(buf, nb)
	//fmt.Println(buf)
	return buf
}