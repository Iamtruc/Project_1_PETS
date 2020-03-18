package main

import (
	"fmt"
	"time"
)

func (io *Input) canEval(cep DummyProtocol)(error, string){
	var err error
	if _, ok := cep.peerCircuit[io.Out]; ok{
		return err, "Input"
	}
	return nil, "Input"
}

func (io *Input) Eval(cep DummyProtocol)(){
	cep.peerCircuit[io.Out] = cep.peerInput[io.Party]
}

func (ao *Add) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[ao.In1]
	_, ok2 := cep.peerCircuit[ao.In2]
	_, not_ok := cep.peerCircuit[ao.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil, "Add"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "Add"
}

func (ao *Add) Eval(cep DummyProtocol){
	cep.peerCircuit[ao.Out] = cep.peerCircuit[ao.In1] + cep.peerCircuit[ao.In2]
}

func (so *Sub) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[so.In1]
	_, ok2 := cep.peerCircuit[so.In2]
	_, not_ok := cep.peerCircuit[so.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil, "Sub"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "Sub"
}

func (so *Sub) Eval(cep DummyProtocol){
	cep.peerCircuit[so.Out] = cep.peerCircuit[so.In1] - cep.peerCircuit[so.In2]
}

func (aco *AddCst) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[aco.In]
	_, not_ok := cep.peerCircuit[aco.Out]
	if (ok1) && (!not_ok){
		return nil, "AddCst"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "Addcst"
}

func (aco *AddCst) Eval(cep DummyProtocol){
	if cep.ID == 0{
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In] + aco.CstValue
	}
	if cep.ID != 0 {
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In]
	}
}

func (mo *Mult) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[mo.In1]
	_, ok2 := cep.peerCircuit[mo.In2]
	_, not_ok := cep.peerCircuit[mo.Out]
	if (ok1) && (ok2) && (!not_ok){
		return nil, "Mult"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "Mult"
}

func (mo *Mult) Eval(cep DummyProtocol){

	var xminusa = cep.peerCircuit[mo.In1] - cep.BeaverA
	var yminusb = cep.peerCircuit[mo.In2] - cep.BeaverB
	var My_c = cep.BeaverA * cep.BeaverC

	xminusa, yminusb = cep.BeaverProt.exchangexminusa(xminusa, yminusb)

	switch cep.ID{
	case 0:
		cep.peerCircuit[mo.Out] = My_c + cep.peerCircuit[mo.In1] * yminusb + cep.peerCircuit[mo.In2] * xminusa -  xminusa * yminusb
	default:
		cep.peerCircuit[mo.Out] = My_c + cep.peerCircuit[mo.In1] * yminusb + cep.peerCircuit[mo.In2] * xminusa
		time.Sleep(time.Second/10)
		}

}

func (mco *MultCst) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[mco.In]
	_, not_ok := cep.peerCircuit[mco.Out]
	if (ok1) && (!not_ok){
		return nil, "MultCst"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "MultCst"
}

func (mco *MultCst) Eval(cep DummyProtocol){
	cep.peerCircuit[mco.Out] = cep.peerCircuit[mco.In] * mco.CstValue
}

func (ro *Reveal) canEval(cep DummyProtocol)(error, string){
	var err error
	_, ok1 := cep.peerCircuit[ro.In]
	_, not_ok := cep.peerCircuit[ro.Out]
	if (ok1) && (!not_ok){
		return nil, "Reveal"
		// Be very careful, empty map arguments are initialized as 0 and problem if you want to add 0
	}
	return err, "Reveal"
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