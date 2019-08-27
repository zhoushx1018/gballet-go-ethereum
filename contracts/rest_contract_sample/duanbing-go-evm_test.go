package main

import (
	"bytes"
	"fmt"
	ec "github.com/duanbing/go-evm/core"
	"github.com/duanbing/go-evm/state"
	"github.com/duanbing/go-evm/vm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/zhoushx1018/gballet-go-ethereum/contracts/rest_contract_sample/infra"
	"math/big"
	"os"
	"testing"
)

// TODO:current test code , is base go-ethereum V1.8.0
//	when this evm package is stable ,need to update to new version, like  V1.8.23

func TestDuanBingGoEvm(t *testing.T) {

	abiFileName := "./testdata/coin_sol_Coin.abi"
	binFileName := "./testdata/coin_sol_Coin.bin"
	data := infra.LoadBin(binFileName)

	msg := ec.NewMessage(infra.FromAddress, &infra.ToAddress, infra.Nonce, infra.Amount, infra.GasLimit, big.NewInt(0), data, false)
	cc := infra.ChainContext{}
	ctx := ec.NewEVMContext(msg, cc.GetHeader(infra.TestHash, 0), cc, &infra.FromAddress)
	dataPath := "/tmp/htdfTmpTestData_duanbing-go-evm"
	os.Remove(dataPath)
	mdb, err := ethdb.NewLDBDatabase(dataPath, 100, 100)
	infra.Must(err)
	db := state.NewDatabase(mdb)

	root := common.Hash{}
	statedb, err := state.New(root, db)
	infra.Must(err)

	//set balance
	statedb.GetOrNewStateObject(infra.FromAddress)
	statedb.GetOrNewStateObject(infra.ToAddress)
	statedb.AddBalance(infra.FromAddress, big.NewInt(1e18))
	testBalance := statedb.GetBalance(infra.FromAddress)
	fmt.Println("init testBalance =", testBalance)
	infra.Must(err)

	//	config := params.TestnetChainConfig
	config := params.MainnetChainConfig
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger /*, JumpTable: vm.NewByzantiumInstructionSet()*/}

	evm := vm.NewEVM(ctx, statedb, config, vmConfig)
	contractRef := vm.AccountRef(infra.FromAddress)
	contractCode, contractAddr, gasLeftover, vmerr := evm.Create(contractRef, data, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)
	//fmt.Printf("getcode:%x\n%x\n", contractCode, statedb.GetCode(contractAddr))

	statedb.SetBalance(infra.FromAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = statedb.GetBalance(infra.FromAddress)
	fmt.Println("after create contract, testBalance =", testBalance)
	abiObj := infra.LoadAbi(abiFileName)

	input, err := abiObj.Pack("minter")
	infra.Must(err)
	outputs, gasLeftover, vmerr := evm.Call(contractRef, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)

	//fmt.Printf("minter is %x\n", common.BytesToAddress(outputs))
	//fmt.Printf("call address %x\n", contractRef)

	sender := common.BytesToAddress(outputs)

	if !bytes.Equal(sender.Bytes(), infra.FromAddress.Bytes()) {
		fmt.Println("caller are not equal to minter!!")
		os.Exit(-1)
	}

	senderAcc := vm.AccountRef(sender)

	input, err = abiObj.Pack("mint", sender, big.NewInt(1000000))
	infra.Must(err)
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)

	// get balance
	input, err = abiObj.Pack("balances", sender)
	infra.Must(err)
	outputs, gasLeftover, vmerr = evm.Call(contractRef, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)
	fmt.Printf("contract balance, after mint|minterAddress=%s|Balance=%x\n", sender.String(), outputs)

	statedb.SetBalance(infra.FromAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = evm.StateDB.GetBalance(infra.FromAddress)

	input, err = abiObj.Pack("send", infra.ToAddress, big.NewInt(11))
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)

	//send
	input, err = abiObj.Pack("send", infra.ToAddress, big.NewInt(19))
	infra.Must(err)
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)

	// get balance
	input, err = abiObj.Pack("balances", infra.ToAddress)
	infra.Must(err)
	outputs, gasLeftover, vmerr = evm.Call(contractRef, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)
	fmt.Printf("contract balance, after send|toAddress=%s|Balance=%x\n", infra.ToAddress.String(), outputs)

	// get balance
	input, err = abiObj.Pack("balances", sender)
	infra.Must(err)
	outputs, gasLeftover, vmerr = evm.Call(contractRef, contractAddr, input, statedb.GetBalance(infra.FromAddress).Uint64(), big.NewInt(0))
	infra.Must(vmerr)
	fmt.Printf("contract balance, after send|minterAddress=%s|Balance=%x\n", sender.String(), outputs)

	// get event
	logs := statedb.Logs()

	for _, log := range logs {
		fmt.Printf("%#v\n", log)
		for _, topic := range log.Topics {
			fmt.Printf("topic: %#v\n", topic)
		}
		fmt.Printf("data: %#v\n", log.Data)
	}

	root, err = statedb.Commit(true)
	infra.Must(err)
	err = db.TrieDB().Commit(root, true)
	infra.Must(err)

	fmt.Println("Root Hash", root.Hex())
	mdb.Close()

	mdb2, err := ethdb.NewLDBDatabase(dataPath, 100, 100)
	infra.Must(err)
	db2 := state.NewDatabase(mdb2)
	statedb2, err := state.New(root, db2)
	infra.Must(err)
	testBalance = statedb2.GetBalance(infra.FromAddress)
	fmt.Println("get testBalance =", testBalance)
	if !bytes.Equal(contractCode, statedb2.GetCode(contractAddr)) {
		fmt.Println("BUG!,the code was changed!")
		os.Exit(-1)
	}
}
