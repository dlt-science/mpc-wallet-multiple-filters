module github.com/taurusgroup/multi-party-sig

go 1.19

require (
	github.com/cronokirby/saferith v0.33.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0
	github.com/fxamacker/cbor/v2 v2.4.0
	github.com/stretchr/testify v1.8.4
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.10.0
	golang.org/x/sync v0.3.0
)

require (
	cuckoo v1.0.0
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	// github.com/dlt-science/crypto-mpc-wallet-bloom v0.0.0-20231024161105-5d7ef9b096ec // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sys v0.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace cuckoo => ./filters/cuckoo
replace (
	cuckoo v1.0.0 => ./filters/cuckoo
)
