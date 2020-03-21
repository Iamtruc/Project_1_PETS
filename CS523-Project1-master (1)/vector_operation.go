package main

import(
	"math/rand"
	)

func newRandomVec(n,T uint64)([]uint64){
	var MyVec []uint64
	var MyLen = int(n)
	for i:=0;i<MyLen;i++{
		MyVec = append(MyVec, uint64(rand.Int63n(int64(T))))
	}
	return MyVec
}

func addVec(a,b []uint64) ([]uint64){
	var MyLen = len(a)
	var MySum []uint64
	for i:=0;i<MyLen;i++{
		MySum = append(MySum, a[i]+b[i])
	}
	return MySum
}

func subVec(a,b []uint64) ([]uint64){
	var MyLen = len(a)
	var MySub []uint64
	for i:=0;i<MyLen;i++{
		MySub = append(MySub, a[i]-b[i])
	}
	return MySub
}

func mulVec(a,b []uint64) ([]uint64){
	var MyLen = len(a)
	var MyMul []uint64
	for i:=0;i<MyLen;i++{
		MyMul = append(MyMul, a[i]*b[i])
	}
	return MyMul
}

func negVec(a []uint64, T uint64) ([]uint64){
	var MyLen = len(a)
	var k uint64
	var MyNeg []uint64
	for i:=0;i<MyLen;i++{
		k = 1
		for k*T < a[i]{
			k++
		}
		MyNeg = append(MyNeg, k*T - a[i])
	}
	return MyNeg
}