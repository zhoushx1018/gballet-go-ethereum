package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/zhoushx1018/gballet-go-ethereum/contracts/rest_contract_sample/infra"

	"os"
	"testing"
)

// TODO:current test code , is base go-ethereum V1.8.0
//	when this evm package is stable ,need to update to new version, like  V1.8.23


func TestEwasmEvm(t *testing.T) {

		dataPath := "/tmp/htdfTmpTestData_ewasmEvm"
		os.Remove(dataPath)
		mdb, err := ethdb.NewLDBDatabase(dataPath, 100, 100)
		infra.Must(err)
		db := state.NewDatabase(mdb)

		root := common.Hash{}
		statedb, err := state.New(root, db)
		infra.Must(err)

		logConfig := vm.LogConfig{}
		structLogger := vm.NewStructLogger(&logConfig)
		vmConfig := vm.Config{Debug: true, Tracer: structLogger /*, JumpTable: vm.NewByzantiumInstructionSet()*/}

		fmt.Printf("statedb=%v|vmconfig=%v\n" , statedb,  vmConfig)

		//var vmTest tests.VMTest
		//
		//vmTest.NewEVM(statedb,vmConfig)

}
