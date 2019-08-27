package main

import (
	"encoding/hex"
	"github.com/gballet/go-ethereum/common"
	"math/big"

	"./infra"
	"fmt"
	"os"
)

var (
	//strTestContractToAddress = "0x70b1151284da341427f6c2dc9b0af63f818fb926"
	strTestContractToAddress = "70b1151284da341427f6c2dc9b0af63f818fb926"
)


func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage:  %s abiFileName  binFilename minterAddress \n", os.Args[0])
		fmt.Printf("    ##minterAddress  : \"nil\" means have no minterAddress\n")
		os.Exit(1)
	}

	abiFileName := os.Args[1]
	binFileName := os.Args[2]
	strMinterAddress := os.Args[3]

	//create contract
	data := infra.LoadBin(binFileName)
	fmt.Printf("contractCode, create contract|Code=%s\n", hex.EncodeToString(data))

	//minter
	abiObj := infra.LoadAbi(abiFileName)
	contractByteCode, err := abiObj.Pack("minter")
	infra.Must(err)
	fmt.Printf("contractCode, minter|Code=%s\n", hex.EncodeToString(contractByteCode))

	//==================access created contract=====================================
	if strMinterAddress == "nil" {
		fmt.Printf("have no strMinterAddress\n")
		os.Exit(0)
	}

	//address convert
	fmt.Printf("strMinterAddress=%s|strTestContractToAddress=%s\n", strMinterAddress, strTestContractToAddress)

	eaMinterAddress := common.HexToAddress(strMinterAddress)
	eaTestContractToAddress := common.HexToAddress(strTestContractToAddress)

	//mint
	contractByteCode, err = abiObj.Pack("mint", eaMinterAddress, big.NewInt(1000000))
	infra.Must(err)
	fmt.Printf("contractCode, mint|strMinterAddress=%s|Code=%s\n", strMinterAddress, hex.EncodeToString(contractByteCode))

	//send
	contractByteCode, err = abiObj.Pack("send", eaTestContractToAddress, big.NewInt(30))
	infra.Must(err)
	fmt.Printf("contractCode, send|strTestContractToAddress=%s|Code=%s\n", strTestContractToAddress, hex.EncodeToString(contractByteCode))

	//get balance
	contractByteCode, err = abiObj.Pack("balances", eaTestContractToAddress)
	infra.Must(err)
	fmt.Printf("contractCode, get balance|strTestContractToAddress=%s|Code=%s\n", strTestContractToAddress, hex.EncodeToString(contractByteCode))

	//get minter balance
	contractByteCode, err = abiObj.Pack("balances", eaMinterAddress)
	infra.Must(err)
	fmt.Printf("contractCode, get balance|strMinterAddress=%s|Code=%s\n", strMinterAddress, hex.EncodeToString(contractByteCode))

}
