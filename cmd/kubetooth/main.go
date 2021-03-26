package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/pa3ng/kubetooth/models"
	k8sclient "github.com/pa3ng/kubetooth/pkg/kubernetes"

	"time"
)

func newBlockchain() *models.Blockchain {
	blockchain := viper.New()
	blockchain.AddConfigPath("./config")
	blockchain.AddConfigPath(".")
	blockchain.SetConfigName("blockchain")
	err := blockchain.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	var blockchainC models.Blockchain
	err = blockchain.Unmarshal(&blockchainC)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return &blockchainC
}

func main() {
	b := newBlockchain()
	fmt.Printf("Deploying %s blockchain version %s running on %s consensus\n",
		b.Ledger,
		b.Version,
		b.Consensus,
	)

	kclient, err := k8sclient.New()
	if err != nil {
		panic(err)
	}

	if err = kclient.CreateKeysConfigMap("keys-config", b.Nodes); err != nil {
		panic(err)
	}

	if err = kclient.CreateService(*b); err != nil {
		panic(err)
	}

	if err := kclient.Deploy(*b); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second) // replace with a check on the validator's pod status/readiness
	if err := kclient.DeployTPs(*b); err != nil {
		panic(err)
	}
}
