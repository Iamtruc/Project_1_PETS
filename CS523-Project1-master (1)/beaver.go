package main

import (
	"math/rand"
)

type BeaverMessage struct {
	a uint64
	b uint64
}


func (cep *DummyProtocol)ask_beaver() (uint64, uint64){
	mod :=10
	a :=  uint64(rand.Intn(mod))
	b :=  uint64(rand.Intn(mod))
	return a,b
}