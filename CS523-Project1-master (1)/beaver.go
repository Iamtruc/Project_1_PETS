package main

import (
	"math/rand"
)

type BeaverMessage struct {
	a uint64
	b uint64
	c uint64
}

func generate_beaver(nb_party int, mod int) ([]uint64, []uint64, uint64){
	var a_list []uint64
	var b_list []uint64
	var sum_a uint64 = 0
	var sum_b uint64 = 0
	for i:= 0; i <nb_party; i++{
		a :=  uint64(rand.Intn(mod))
		b :=  uint64(rand.Intn(mod))
		a_list = append(a_list, a)
		b_list = append(b_list, b)
		sum_a += a
		sum_b +=b
	}
	return a_list, b_list, sum_a * sum_b
}

func (cep *DummyProtocol)ask_beaver() {
	mod :=10
	n := len(cep.Peers)
	a, b, c := generate_beaver(n, mod)
	for i, peer := range cep.Peers {
		peer.Chan_beav <- BeaverMessage{a[i], b[i], c}

	}
}