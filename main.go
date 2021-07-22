package main

import (
	"log"
	"os"
	"strconv"
	ethereum "swap.io-agent/src/blockchain/ethereum/blockchainIndexer"
	"swap.io-agent/src/configLoader"
	"swap.io-agent/src/httpHandler"
	"swap.io-agent/src/httpServer"
	"swap.io-agent/src/levelDbStore"
	"swap.io-agent/src/redisStore"
	"swap.io-agent/src/serviceRegistry"
	"swap.io-agent/src/socketServer"
)

func main() {
	registry := serviceRegistry.NewServiceRegistry()

	err := configLoader.InitializeConfig()
	if err != nil {panic(err)}

	db, err := redisStore.InitializeDB()
	if err != nil {
		log.Panicf("redisStore not initialize, err: %v", err)
	}
	err = registry.RegisterService(&db)
	if err != nil {
		log.Panicln(err.Error())
	}

	blockchainName := os.Getenv("BLOCKCHAIN")
	if len(blockchainName) == 0 {
		log.Panicln("SET BLOCKCHAIN SETTINGS IN ENV")
	}
	blockchainDefaultScannedBlock, err := strconv.Atoi(
		os.Getenv("BLOCKCHAIN_DEFAULT_SCANNED_BLOCK"),
	)
	if err != nil {
		log.Panicln("SET BLOCKCHAIN_DEFAULT_SCANNED_BLOCK SETTINGS NUM IN ENV")
	}
	transactionStore, err := levelDbStore.InitialiseTransactionStore(
		levelDbStore.TransactionsStoreConfig{
			Name: blockchainName,
			DefaultScannedBlocks: blockchainDefaultScannedBlock,
		},
	)
	if err != nil {
		log.Panicln(err.Error())
	}
	err = registry.RegisterService(transactionStore)
	if err != nil {
		log.Panicln(err.Error())
	}

	ethereum.BlockchainIndexerRegister(registry)

	socketServer.Register(registry)

	httpServerEntity := httpServer.InitializeServer()
	err = registry.RegisterService(httpServerEntity)
	if err != nil {
		log.Panicln(err.Error())
	}

	httpHandlerEntity := httpHandler.InitializeServer()
	err = registry.RegisterService(httpHandlerEntity)
	if err != nil {
		log.Panicln(err.Error())
	}

	registry.StartAll()

	<-make(chan struct{})
}