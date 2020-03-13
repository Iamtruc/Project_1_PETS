package main

import (
	"fmt"
	"time"
)

func (io *Input) canEval(cep DummyProtocol)(error){
	var err error
	if _, ok := cep.peerCircuit[io.Out]; ok{
		return err
	}
	return nil
}

func (io *Input) Eval(cep DummyProtocol)(){
	cep.peerCircuit[io.Out] = cep.peerInput[io.Party]
}

func (ao *Add) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[ao.In1]
	_, ok2 := cep.peerCircuit[ao.In2]
	_, not_ok := cep.peerCircuit[ao.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (ao *Add) Eval(cep DummyProtocol){
	cep.peerCircuit[ao.Out] = cep.peerCircuit[ao.In1] + cep.peerCircuit[ao.In2]
}

func (so *Sub) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[so.In1]
	_, ok2 := cep.peerCircuit[so.In2]
	_, not_ok := cep.peerCircuit[so.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (so *Sub) Eval(cep DummyProtocol){
	cep.peerCircuit[so.Out] = cep.peerCircuit[so.In1] - cep.peerCircuit[so.In2]
}

func (aco *AddCst) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[aco.In]
	_, not_ok := cep.peerCircuit[aco.Out]
	if (ok1) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (aco *AddCst) Eval(cep DummyProtocol){
	if cep.ID == 0{
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In] + aco.CstValue
	}
	if cep.ID != 0 {
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In]
	}
}

func (mo *Mult) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[mo.In1]
	_, ok2 := cep.peerCircuit[mo.In2]
	_, not_ok := cep.peerCircuit[mo.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (mo *Mult) Eval(cep DummyProtocol){
	var a uint64
	var b uint64
	var c uint64
	var n = len(cep.Peers)

	if cep.ID == 0{

	}

	time.Sleep(time.Second/100)

	// I wanted to create a Beaver_Channel which would send only the values x-a and y-b but I failed to do so.
	// I decided to do something more simple : if the party indice is -1 : you send x-a, if it is -2, you send y-b
	xminusa := cep.peerCircuit[mo.In1] - a
	yminusb := cep.peerCircuit[mo.In2] - b
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan<- DummyMessage{PartyID(n + 1), xminusa}
			peer.Chan<- DummyMessage{PartyID(n + 2), yminusb}
		}
	}

	received := 0
	for m := range cep.Chan {

		if m.Party == PartyID(n + 1){
			xminusa += m.Value
			received ++
		}
		if m.Party == PartyID(n + 2){
			yminusb += m.Value
			received ++
		}
		if received == 2*(len(cep.Peers)-1) {
			break
		}
	}

	cep.peerCircuit[mo.Out] = c + cep.peerCircuit[mo.In1] * yminusb + cep.peerCircuit[mo.In2] * xminusa
	if cep.ID == 0{
		cep.peerCircuit[mo.Out]-= xminusa * yminusb
	}
	time.Sleep(time.Second/100) // Just to make sure that everybody is on the same page. Problem : slow.
}

func (mco *MultCst) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[mco.In]
	_, not_ok := cep.peerCircuit[mco.Out]
	if (ok1) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (mco *MultCst) Eval(cep DummyProtocol){
	cep.peerCircuit[mco.Out] = cep.peerCircuit[mco.In] * mco.CstValue
}

func (ro *Reveal) canEval(cep DummyProtocol)(error){
	var err error
	_, ok1 := cep.peerCircuit[ro.In]
	_, not_ok := cep.peerCircuit[ro.Out]
	if (ok1) && (!not_ok){
		return nil
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err
}

func (ro *Reveal) Eval(cep DummyProtocol){
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan <- DummyMessage{cep.ID, cep.peerCircuit[ro.In]}
		}
	}
	cep.peerCircuit[ro.Out] = cep.peerCircuit[ro.In]

	received := 0
	for m := range cep.Chan {
		cep.peerCircuit[ro.Out] += m.Value
		received++
		if received == len(cep.Peers)-1 {
			cep.Output = cep.peerCircuit[ro.Out]
			fmt.Println(cep.ID, "completed with output", cep.Output)
			break
		}
	}

	if cep.WaitGroup != nil {
		cep.WaitGroup.Done()
	}
}