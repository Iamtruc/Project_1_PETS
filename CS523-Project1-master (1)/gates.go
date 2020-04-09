package main

import "time"

// We define three methods for each operation : Identify which indicates what is the nature of the operation
//The method "canEval" which checks that the operation will not lead to an error
// We then have the operation "Eval" which carries out the operation.

func (io *Input) Identify()(string){
	return "Input"
}

func (io *Input) canEval(cep DummyProtocol)(error, string){
	var err error
	if _, ok := cep.peerCircuit[io.Out]; ok{
		return err, "Input"
	}
	return nil, "Input"
}

func (io *Input) Eval(cep DummyProtocol)(uint64){
	cep.peerCircuit[io.Out] = cep.peerInput[io.Party]
	return cep.peerCircuit[io.Out]
}

func (ao *Add) Identify()(string){
	return "Add"
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

func (ao *Add) Eval(cep DummyProtocol)(uint64){
	cep.peerCircuit[ao.Out] = cep.peerCircuit[ao.In1] + cep.peerCircuit[ao.In2]
	return cep.peerCircuit[ao.Out]
}

func (so *Sub) Identify()(string){
	return "Sub"
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

func (so *Sub) Eval(cep DummyProtocol)(uint64){
	cep.peerCircuit[so.Out] = cep.peerCircuit[so.In1] - cep.peerCircuit[so.In2]
	return cep.peerCircuit[so.Out]
}

func (aco *AddCst) Identify()(string){
	return "AddCst"
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

func (aco *AddCst) Eval(cep DummyProtocol)(uint64){
	// Only one of the peers should add the constant, or else the result will be off. We choose that the operations are carried out by the peer whose ID is 0
	if cep.ID == 0{
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In] + aco.CstValue
	}
	if cep.ID != 0 {
		cep.peerCircuit[aco.Out] = cep.peerCircuit[aco.In]
	}
	return cep.peerCircuit[aco.Out]
}

func (mo *Mult) Identify()(string){
	return "Mult"
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

func (mo *Mult) Eval(cep DummyProtocol)(uint64){
	// The multiplication is slightly more complicated, it demands to exchange the values with the Beaver Triplets

	var xminusa = cep.peerCircuit[mo.In1] - cep.BeaverA
	var yminusb = cep.peerCircuit[mo.In2] - cep.BeaverB
	var My_c = cep.BeaverC

	xminusa, yminusb = cep.BeaverProt.exchangexminusa(xminusa, yminusb)

	switch cep.ID{
	case 0:
		cep.peerCircuit[mo.Out] = My_c + cep.peerCircuit[mo.In1] * yminusb + cep.peerCircuit[mo.In2] * xminusa -  xminusa * yminusb
	default:
		cep.peerCircuit[mo.Out] = My_c + cep.peerCircuit[mo.In1] * yminusb + cep.peerCircuit[mo.In2] * xminusa
		time.Sleep(time.Second/10)
		}
		return cep.peerCircuit[mo.Out]
}

func (mco *MultCst) Identify()(string){
	return "MultCst"
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

func (mco *MultCst) Eval(cep DummyProtocol)(uint64){
	cep.peerCircuit[mco.Out] = cep.peerCircuit[mco.In] * mco.CstValue
	return cep.peerCircuit[mco.Out]
}

func (ro *Reveal) Identify()(string){
	return "Reveal"
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

func (ro *Reveal) Eval(cep DummyProtocol) (uint64){
	// In Reveal, the peers exchange their results and reconstruct the total. They then close the circuit.
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
			close(cep.Chan)
		}
	}
	return cep.peerCircuit[ro.Out]
}