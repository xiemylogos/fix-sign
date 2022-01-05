package main

import (
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
	"github.com/ontio/ontology/events"
	"github.com/urfave/cli"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "supply sign CLI"
	app.Action = startSupplySign
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{}
	app.Flags = []cli.Flag{
		utils.ConfigFlag,
		utils.NetworkIdFlag,
		utils.ReservedPeersOnlyFlag,
		utils.ReservedPeersFileFlag,
	}
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
	cfg := initNodeConfig(ctx)
	ldg, err := initLedger(ctx, cfg)
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

func initNodeConfig(ctx *cli.Context) *config.OntologyConfig {
	cfg, err := cmd.SetOntologyConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to process node config: %s", err)
	}
	return cfg
}

func initLedger(ctx *cli.Context, cfg *config.OntologyConfig) (*ledger.Ledger, error) {
	events.Init()
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		log.Errorf("GetBookkeepers error: %s", err)
		return nil, err
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		log.Errorf("genesisBlock error %s", err)
	}
	ldg, err := ledger.InitLedger(cfg.Common.DataDir, config.GetStateHashCheckHeight(cfg.P2PNode.NetworkId), bookKeepers, genesisBlock)
	if err != nil {
		log.Fatalf("Failed to open ledger: %s", err)
		return nil, err
	}
	SetExitHandler(func() {
		log.Info("Closing ledger")
		if err := ldg.Close(); err != nil {
			log.Errorf("Failed to close ledger: %s", err)
		}
	})
	log.Info("Ledger init success")
	ledger.DefLedger = ldg
	return ldg, nil
}
