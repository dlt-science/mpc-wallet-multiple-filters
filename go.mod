module github.com/dlt-science/mpc-wallet-multiple-filters

go 1.19

require (
	github.com/cronokirby/saferith v0.33.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0
	github.com/fxamacker/cbor/v2 v2.4.0
	github.com/stretchr/testify v1.8.4
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.17.0
	golang.org/x/sync v0.5.0
)

require (
	cuckoo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/btcsuite/btcd v0.24.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.7.0 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20200812162917-85c65e2d0165 // indirect
	github.com/dlt-science/mpc-wallet-multiple-filters/filters/cuckoo v0.0.0-20240103142535-5fa084de62c2 // indirect
	github.com/ethereum/c-kzg-4844 v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/irfansharif/cfilter v0.1.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/panmari/cuckoofilter v1.0.6 // indirect
	github.com/seiflotfy/cuckoofilter v0.0.0-20220411075957-e3b120b3f5fb // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

require (
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.6.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.13.8
	 github.com/dlt-science/crypto-mpc-wallet-bloom v0.0.0-20231024161105-5d7ef9b096ec // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sys v0.15.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace cuckoo => ./filters/cuckoo
replace cuckoo v1.0.0 => ./filters/cuckoo
