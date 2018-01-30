package siatest

import (
	"bytes"
	"errors"
	"unsafe"

	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/encoding"
	"github.com/NebulousLabs/Sia/node"
	"github.com/NebulousLabs/Sia/node/api/client"
	"github.com/NebulousLabs/Sia/node/api/server"
	"github.com/NebulousLabs/Sia/types"
)

// TestNode is a helper struct for testing that contains a server and a client
// as embedded fields.
type TestNode struct {
	server.Server
	client.Client
	primarySeed string
}

// NewNode creates a new funded TestNode
func NewNode(nodeParams node.NodeParams) (*TestNode, error) {
	address := ":9980"
	userAgent := "Sia-Agent"
	password := "password"

	// We can't create a funded node without a miner
	if !nodeParams.CreateMiner && nodeParams.Miner == nil {
		return nil, errors.New("Can't create funded node without miner")
	}

	// Create client
	c := client.New(address)
	c.UserAgent = userAgent
	c.Password = password

	// Create server
	s, err := server.New(address, userAgent, password, nodeParams)
	if err != nil {
		return nil, err
	}

	// Create TestNode
	tn := &TestNode{*s, *c, ""}

	// Init wallet
	wip, err := tn.PostWalletInit("", false)
	if err != nil {
		return nil, err
	}
	tn.primarySeed = wip.PrimarySeed

	// Unlock wallet
	if err := tn.PostWalletUnlock(tn.primarySeed); err != nil {
		return nil, err
	}

	// fund the node
	for i := types.BlockHeight(0); i <= types.MaturityDelay; i++ {
		if err := tn.MineBlock(); err != nil {
			return nil, err
		}
	}

	// Return TestNode
	return tn, nil
}

// MineBlock makes the underlying node mine a single block and broadcast it.
func (tn *TestNode) MineBlock() error {
	// Get the header
	target, header, err := tn.GetMinerHeader()
	if err != nil {
		return build.ExtendErr("failed to get header for work", err)
	}
	// Solve the header
	solveHeader(target, header)

	// Submit the header
	if err := tn.PostMinerHeader(header); err != nil {
		return build.ExtendErr("failed to submit header", err)
	}
	return nil
}

// solveHeader solves the header by finding a nonce for the target
func solveHeader(target types.Target, bh types.BlockHeader) error {
	header := encoding.Marshal(bh)

	// try 16e3 times to solve block
	var nonce uint64
	for i := 0; i < 16e3; i++ {
		id := crypto.HashBytes(header)
		if bytes.Compare(target[:], id[:]) >= 0 {
			copy(bh.Nonce[:], header[32:40])
			return nil
		}
		*(*uint64)(unsafe.Pointer(&header[32])) = nonce
		nonce++
	}
	return errors.New("couldn't solve block")
}