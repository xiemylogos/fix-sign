package main

import (
	"fmt"
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/cmd"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "supply sign CLI"
	app.Action = startSupplySign
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}

func startSupplySign(ctx *cli.Context) {
	ontSdk := sdk.NewOntologySdk()
	accounts, err := LoadAccount(ontSdk)
	if err != nil {
		panic(err)
	}
	ldg, err := initLedger(ctx, 0)
	if err != nil {
		log.Errorf(" initLedger err:%s", err)
		return
	}
	var height uint32
	for height = 0; height < ldg.GetCurrentBlockHeight(); height++ {
		header, err := ldg.GetHeaderByHeight(height)
		if err != nil {
			log.Errorf("GetHeaderByHeight height:%d,err:%s", height, err)
			return
		}
		usedPubKey := make(map[string]bool)
		for _, bookkeeper := range header.Bookkeepers {
			pubkey := vconfig.PubkeyID(bookkeeper)
			usedPubKey[pubkey] = true
		}
		blkHash := header.Hash()
		sigData := make([][]byte, 0)
		bookKeepers := make([]keypair.PublicKey, 0)
		for _, acc := range accounts {
			if !usedPubKey[vconfig.PubkeyID(acc.PublicKey)] {
				sig, err := acc.Sign(blkHash[:])
				if err != nil {
					log.Errorf("sign err:%s,height:%d", err, height)
					return
				}
				sigData = append(sigData, sig)
				bookKeepers = append(bookKeepers, acc.PublicKey)
			}
		}
		headers := make([]*types.Header, 0)
		header.Bookkeepers = bookKeepers
		header.SigData = sigData
		headers = append(headers, header)
		err = ldg.AddHeaders(headers)
		if err != nil {
			log.Errorf("addHeaders err:%s,height:%d", err, height)
			return
		}
		log.Infof("modify header sign height:%d", height)
	}
}

func initLedger(ctx *cli.Context, stateHashHeight uint32) (*ledger.Ledger, error) {
	var err error
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error: %s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	ledger.DefLedger, err = ledger.InitLedger(dbDir, stateHashHeight, bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("NewLedger error: %s", err)
	}
	log.Infof("Ledger init success")
	return ledger.DefLedger, nil
}
