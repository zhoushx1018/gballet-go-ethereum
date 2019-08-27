package infra

import (
	"fmt"
	"github.com/duanbing/go-evm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/zhoushx1018/gballet-go-ethereum/crypto"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

var (
	TestHash    = common.StringToHash("zhoushx")
	FromAddress = common.StringToAddress("UserA")
	ToAddress   = common.StringToAddress("UserB")
	Amount      = big.NewInt(0)
	Nonce       = uint64(0)
	GasLimit    = big.NewInt(100000000000000000)
	Coinbase    = FromAddress
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
func LoadBin(filename string) []byte {
	code, err := ioutil.ReadFile(filename)
	Must(err)
	return hexutil.MustDecode("0x" + string(code))
}
func LoadAbi(filename string) abi.ABI {
	abiFile, err := os.Open(filename)
	Must(err)
	defer abiFile.Close()
	abiObj, err := abi.JSON(abiFile)
	Must(err)
	return abiObj
}

func Print(outputs []byte, name string) {
	fmt.Printf("method=%s, output=%x\n", name, outputs)
}

type ChainContext struct{}

//
//func(cc ChainContext) Engine() consensus.Engine{
//
//}

func (cc ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {

	return &types.Header{
		// ParentHash: common.Hash{},
		// UncleHash:  common.Hash{},
		Coinbase: FromAddress,
		//	Root:        common.Hash{},
		//	TxHash:      common.Hash{},
		//	ReceiptHash: common.Hash{},
		//	Bloom:      types.BytesToBloom([]byte("duanbing")),
		Difficulty: big.NewInt(1),
		Number:     big.NewInt(1),
		GasLimit:   1000000,
		GasUsed:    0,
		Time:       big.NewInt(time.Now().Unix()),
		Extra:      nil,
		//MixDigest:  TestHash,
		//Nonce:      types.EncodeNonce(1),
	}
}

func VmTestBlockHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
