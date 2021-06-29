package oracle

import (
	"crypto/ecdsa"
	"hash"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"golang.org/x/crypto/sha3"
)

func TestWrapGetLatestBlockNumberFn(t *testing.T) {
	key, _ := crypto.GenerateKey()
	sim, db := newSimulatedBackend(key)
	chain := sim.Blockchain()

	getLatest := wrapGetLatestBlockNumberFn(sim)

	// Generate a valid chain of 10 blocks
	blocks, _ := core.GenerateChain(chain.Config(), chain.CurrentBlock(), chain.Engine(), db, 10, nil)

	// Check that the latest is 0 to start
	latest, err := getLatest()
	if err != nil {
		t.Fatal(err)
	}
	if latest != 0 {
		t.Fatal("not zero")
	}

	// Insert the blocks one by one and assert that they are incrementing
	for i, block := range blocks {
		if _, err := chain.InsertChain([]*types.Block{block}); err != nil {
			t.Fatal(err)
		}
		latest, err := getLatest()
		if err != nil {
			t.Fatal(err)
		}
		// Handle zero index by adding 1
		if latest != uint64(i+1) {
			t.Fatal("mismatch")
		}
	}
}

func newSimulatedBackend(key *ecdsa.PrivateKey) (*backends.SimulatedBackend, ethdb.Database) {
	var gasLimit uint64 = 9_000_000
	auth, _ := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	genAlloc := make(core.GenesisAlloc)
	genAlloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(9223372036854775807)}
	db := rawdb.NewMemoryDatabase()
	sim := backends.NewSimulatedBackendWithDatabase(db, genAlloc, gasLimit)
	return sim, db
}

// testHasher is the helper tool for transaction/receipt list hashing.
// The original hasher is trie, in order to get rid of import cycle,
// use the testing hasher instead.
type testHasher struct {
	hasher hash.Hash
}

func newHasher() *testHasher {
	return &testHasher{hasher: sha3.NewLegacyKeccak256()}
}

func (h *testHasher) Reset() {
	h.hasher.Reset()
}

func (h *testHasher) Update(key, val []byte) {
	h.hasher.Write(key)
	h.hasher.Write(val)
}

func (h *testHasher) Hash() common.Hash {
	return common.BytesToHash(h.hasher.Sum(nil))
}
