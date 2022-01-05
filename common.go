package main

import (
	"encoding/json"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"io/ioutil"
)

func GetAccountByPassword(sdk *sdk.OntologySdk, path string) (*sdk.Account, bool) {
	wallet, err := sdk.OpenWallet(path)
	if err != nil {
		log.Error("open wallet error:", err)
		return nil, false
	}
	pwd, err := password.GetPassword()
	if err != nil {
		log.Error("getPassword error:", err)
		return nil, false
	}
	user, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		log.Error("getDefaultAccount error:", err)
		return nil, false
	}
	return user, true
}

type ConfigParam struct {
	Path []string
}

func LoadAccount(ontSdk *sdk.OntologySdk) ([]*sdk.Account, error) {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Errorf("ioutil.ReadFile failed ", err)
		return nil, err
	}
	configParam := new(ConfigParam)
	err = json.Unmarshal(data, configParam)
	if err != nil {
		log.Error("json.Unmarshal failed ", err)
		return nil, err
	}
	var accs []*sdk.Account
	for _, path := range configParam.Path {
		user, ok := GetAccountByPassword(ontSdk, path)
		if !ok {
			return nil, err
		}
		accs = append(accs, user)
	}
	return accs, nil
}
