package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	fmt.Println("Starting")

	alice := sha256.Sum256([]byte("Alice"))
	bob := sha256.Sum256([]byte("Bob"))

	bc, err := NewBlockchain(alice)
	if err != nil {
		panic(err)
	}
	defer bc.Close()

	bc.GenerateUTxO()

	bc.MineNext()

	fmt.Println("Alice has: " + fmt.Sprint(bc.GetUTxOsForUser(alice).Balance()))
	fmt.Println("Bob has: " + fmt.Sprint(bc.GetUTxOsForUser(bob).Balance()))

	if err = bc.Send(alice, bob, 30); err != nil {
		panic(err)
	}
	bc.MineNext()

	fmt.Println("Alice has: " + fmt.Sprint(bc.GetUTxOsForUser(alice).Balance()))
	fmt.Println("Bob has: " + fmt.Sprint(bc.GetUTxOsForUser(bob).Balance()))
}
