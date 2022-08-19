module github.com/Fantom-foundation/substate-cli

go 1.14

require (
	github.com/Fantom-foundation/go-opera v1.1.1-rc.2
	github.com/ethereum/go-ethereum v1.10.8
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/ethereum/go-ethereum => github.com/Fantom-foundation/go-ethereum-substate v1.0.0

replace github.com/Fantom-foundation/go-opera => github.com/Fantom-foundation/go-opera-substate v1.0.0

replace github.com/dvyukov/go-fuzz => github.com/guzenok/go-fuzz v0.0.0-20210103140116-f9104dfb626f
