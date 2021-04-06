package main

import (
	"fmt"
)

func main() {
	fmt.Println("Starting")

	miner, _ := NewAccount()

	bc, err := NewBlockchain(miner)
	if err != nil {
		panic(err)
	}
	defer bc.Close()

	alice, _ := NewAccount()
	bob, _ := NewAccount()
	bc.AddKey(alice.PublicKey)
	bc.AddKey(bob.PublicKey)

	fmt.Println("Alice has: " + fmt.Sprint(bc.GetUTxOsForUser(alice.Id).Balance()))
	fmt.Println("Bob has: " + fmt.Sprint(bc.GetUTxOsForUser(bob.Id).Balance()))
	fmt.Println("The miner has: " + fmt.Sprint(bc.GetUTxOsForUser(miner.Id).Balance()))

	if err = bc.Send(miner, alice.Id, 50); err != nil {
		panic(err)
	}
	bc.MineNext()

	fmt.Println("Alice has: " + fmt.Sprint(bc.GetUTxOsForUser(alice.Id).Balance()))
	fmt.Println("Bob has: " + fmt.Sprint(bc.GetUTxOsForUser(bob.Id).Balance()))
	fmt.Println("The miner has: " + fmt.Sprint(bc.GetUTxOsForUser(miner.Id).Balance()))

	if err = bc.Send(alice, bob.Id, 30); err != nil {
		panic(err)
	}

	block, err := bc.MineNext()
	if err != nil {
		panic(err)
	}
	err = bc.VerifyBlock(block)
	if err != nil {
		panic(err)
	}

	fmt.Println("Alice has: " + fmt.Sprint(bc.GetUTxOsForUser(alice.Id).Balance()))
	fmt.Println("Bob has: " + fmt.Sprint(bc.GetUTxOsForUser(bob.Id).Balance()))
	fmt.Println("The miner has: " + fmt.Sprint(bc.GetUTxOsForUser(miner.Id).Balance()))
}
