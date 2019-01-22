package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

func check(e error) {
	if e != nil {
		panic(fmt.Sprintf("💩 %v", e))
	}
}

type ConflictTracer struct {
	Errors uint64
	Writes map[string]uint64
	Reads  map[string]uint64
}

func newConflictTracer() *ConflictTracer {
	return &ConflictTracer{
		Writes: make(map[string]uint64),
		Reads:  make(map[string]uint64),
	}
}

// PrecompileTracer is a helper for finding calls to precompiles
type PrecompileTracer struct {
	from     *common.Address
	to       *common.Address
	blockNum *big.Int
	txNum    uint64

	cf *ConflictTracer
	// locs    []string
	// address []string
	// write   []bool
}

func newPrecompileTracer(blockNum uint64, txNum uint64, cf *ConflictTracer) *PrecompileTracer {
	return &PrecompileTracer{
		blockNum: big.NewInt(int64(blockNum)),
		txNum:    txNum,
		cf:       cf,
		// locs:     make([]string, 0),
		// address:  make([]string, 0),
		// write:    make([]bool, 0),
	}
}

// CaptureStart
func (pt PrecompileTracer) CaptureStart(from common.Address, to common.Address, call bool, input []byte, gas uint64, value *big.Int) error {
	pt.from = &from
	pt.to = &to
	return nil
}

// CaptureState
func (pt PrecompileTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, contract *vm.Contract, depth int, err error) error {
	switch {
	case op == vm.CALL || op == vm.CALLCODE || op == vm.STATICCALL || op == vm.DELEGATECALL:
		addr := stack.Back(1).Uint64()
		if addr == 1 || addr == 5 {
			fmt.Println(op, "🙋 CALL precompile addr=", addr, "pc=", pc, "contract addr=", contract.Address().Hex(), "gas=", gas, "depth=", depth, "tx #", pt.txNum, "block #", pt.blockNum)
		}
		break
		// case op == vm.SSTORE || op == vm.SLOAD:
		// 	addr := fmt.Sprintf("%#x", stack.Back(0))
		// 	caddr := contract.Address().Hex()
		// pt.locs = append(pt.locs, addr)
		// pt.address = append(pt.address, caddr)
		// pt.write = append(pt.write, op == vm.SSTORE)
		// break
	default:
		break
	}
	return nil
}

// CaptureFault
func (pt PrecompileTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, contract *vm.Contract, depth int, err error) error {
	pt.cf.Errors++
	return nil
}

// CaptureEnd
func (pt PrecompileTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	// for i, write := range pt.write {
	// 	key := fmt.Sprintf("%s:%s", pt.address[i], pt.locs[i])
	// 	if write {
	// 		pt.cf.Writes[key]++
	// 	} else {
	// 		pt.cf.Reads[key]++
	// 	}
	// }
	return nil
}
func (pt PrecompileTracer) CaptureCreate(creator common.Address, creation common.Address) error {
	return nil
}

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Sprintf("💣 Need start block as args, got: %v", os.Args))
	}

	var fromBlock int
	if fb, err := strconv.Atoi(os.Args[1]); err != nil {
		check(err)
	} else {
		fromBlock = fb
	}

	fmt.Println("Opening DB", time.Now())
	ethDb, err := ethdb.NewLDBDatabase("/opt/chains/tracing/geth/chaindata", 0, 0)
	fmt.Println("Database opened", time.Now())
	check(err)
	defer ethDb.Close()
	bc, err := core.NewBlockChain(ethDb, nil, params.MainnetChainConfig, ethash.NewFaker(), vm.Config{}, nil)
	fmt.Println("Blockchain created", time.Now())
	check(err)
	currentBlock := bc.CurrentBlock()
	currentBlockNr := currentBlock.NumberU64()
	fmt.Printf("💬 Current block number: %d\n", currentBlockNr)

	// fw, err := os.OpenFile(fmt.Sprintf("writes_%s.json", os.Args[1]), os.O_CREATE|os.O_WRONLY, 0644)
	// check(err)
	// defer fw.Close()
	// fr, err := os.OpenFile(fmt.Sprintf("reads_%s.json", os.Args[1]), os.O_CREATE|os.O_WRONLY, 0644)
	// check(err)
	// defer fr.Close()

	ct := newConflictTracer()

	parent := bc.GetBlockByNumber(uint64(fromBlock - 1))
	fmt.Println("Got block number", time.Now())

	fmt.Println("🔬", fromBlock, time.Now())
	for i := uint64(fromBlock); i < uint64(fromBlock+1000); i++ {
		statedb, err := bc.StateAt(parent.Root())
		if i%100 == 0 {
			fmt.Println("Got state root", time.Now())
		}
		check(err)
		block := bc.GetBlockByNumber(i)
		signer := types.MakeSigner(params.MainnetChainConfig, block.Number())

		for j, tx := range block.Transactions() {
			vmCfg := vm.Config{Debug: true, Tracer: newPrecompileTracer(i, uint64(j), ct)}
			msg, _ := tx.AsMessage(signer)
			vmctx := core.NewEVMContext(msg, block.Header(), bc, nil)
			vmenv := vm.NewEVM(vmctx, statedb, params.MainnetChainConfig, vmCfg)

			if _, _, _, err := core.ApplyMessage(vmenv, msg, new(core.GasPool).AddGas(msg.Gas())); err != nil {
				check(fmt.Errorf("%v: Error at tx #%d block #%d: %v", time.Now(), j, i, err))
				break
			}
		}

		parent = block
	}
	fmt.Println("🎉", fromBlock+999, time.Now())

	// writes, err := json.Marshal(ct.Writes)
	// check(err)
	// fw.Write(writes)

	// reads, err := json.Marshal(ct.Reads)
	// check(err)
	// fw.Write(reads)
}
