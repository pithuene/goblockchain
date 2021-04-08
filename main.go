package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	//"net/http"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "goblockchain",
		Usage: "Interface for running a blockchain node",
		Action: func(c *cli.Context) error {
			start("blockchain.db", "account")
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func start(dbFile string, accountFile string) {
	fmt.Println("Starting")

	var miner *Account

	if accRaw, err := os.ReadFile(accountFile); err != nil {
		fmt.Println("No account found")
		if miner, err = NewAccount(); err != nil {
			panic(err)
		} else {
			fmt.Println("Generated a new account")
			os.WriteFile(accountFile, miner.Serialize(), 0644)
		}
	} else {
		miner = AccountDeserialize(accRaw)
		fmt.Printf("Account '%x' opened from file\n", miner.Id)
	}

	bc, err := NewBlockchain(dbFile, miner)
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
