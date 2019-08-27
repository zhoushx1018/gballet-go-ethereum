



all:build


build:
	GO111MODULE=off; go run byte_code_sample.go ./testdata/coin_sol_Coin.abi ./testdata/coin_sol_Coin.bin  e7abd8f8c4cc0a2acf0872cf490581db27255f8f
	

test:
	GO111MODULE=off; go test eth_evm_test.go
