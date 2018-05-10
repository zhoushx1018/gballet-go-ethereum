// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package whisperv6

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	set "gopkg.in/fatih/set.v0"
)

// LibP2PWhisperServer implements WhisperServer for libp2p.
type LibP2PWhisperServer struct {
	Host host.Host // Libp2p host structure

	PeerMutex sync.RWMutex  // Guard the list of active peers
	Peers     []*LibP2PPeer // List of active peers

	whisper *Whisper // Pointer to the whisper object
}

func (server *LibP2PWhisperServer) connectToPeer(p *LibP2PPeer) error {
	log.Info("opening stream to peer: ", p.id.Pretty(), "from peer", server.Host.ID().Pretty())

	// Create a stream with the peer
	s, err := server.Host.NewStream(context.Background(), p.id, WhisperProtocolString)
	if err != nil {
		return err
	}

	// Save the stream
	lps := newLibp2pStream(server, s).(*LibP2PStream)

	p.connectionStream = lps
	p.ws = p.connectionStream

	// TODO send my known list of peers

	// Call HandlePeer to perform the handshake
	go server.whisper.HandlePeer(p, p.connectionStream)

	return err
}

// Start starts the server
func (server *LibP2PWhisperServer) Start() error {
	server.Host.SetStreamHandler(WhisperProtocolString, func(stream inet.Stream) {
		log.Info("opening stream from new peer")

		pid := stream.Conn().RemotePeer()
		var peer Peer
		server.PeerMutex.RLock()
		for _, p := range server.Peers {
			if p.id == pid {
				peer = p
				break
			}
		}
		server.PeerMutex.RUnlock()

		lps := newLibp2pStream(server, stream).(*LibP2PStream)

		// Unknown peer
		if peer == nil {
			peer = newLibP2PPeer(server, server.whisper, pid, lps)
			server.PeerMutex.Lock()
			server.Peers = append(server.Peers, peer.(*LibP2PPeer))
			server.PeerMutex.Unlock()
		}

		go server.whisper.HandlePeer(peer, lps)
	})

	fmt.Println("Currently having the following peers:", server.Peers)

	// Open a stream to every peer currently known
	var err error
	server.PeerMutex.RLock()
	for _, p := range server.Peers {
		if e := server.connectToPeer(p); e != nil {
			err = e
		}
	}
	server.PeerMutex.RUnlock()

	return err
}

// Stop stops the server
func (server *LibP2PWhisperServer) Stop() {
	server.PeerMutex.RLock()
	for _, p := range server.Peers {
		// TODO send disconnect message
		p.connectionStream.lp2pStream.Close()
	}
	server.PeerMutex.RUnlock()

	server.Host.Close()
}

// PeerCount returns the peer count for the node
func (server *LibP2PWhisperServer) PeerCount() int {
	server.PeerMutex.RLock()
	defer server.PeerMutex.RUnlock()
	return len(server.Peers)
}

// Enode returns the enode address of the node
func (server *LibP2PWhisperServer) Enode() string {
	addr := server.Host.Addrs()[0]
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", server.Host.ID().Pretty()))
	fullAddr := addr.Encapsulate(hostAddr)
	return fullAddr.String()
}

// AddPeer is a helper function to add peers to the server
func (server *LibP2PWhisperServer) AddPeer(addr ma.Multiaddr) *LibP2PPeer {
	log.Info("Adding peer: ", addr)
	pid, err := addr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		// XXX
		panic(err)
	}
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		panic(err)
	}
	ipfsaddrpart, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))
	ipaddr := addr.Decapsulate(ipfsaddrpart)
	server.Host.Peerstore().AddAddr(peerid, ipaddr, pstore.PermanentAddrTTL)
	newPeer := newLibP2PPeer(server, server.whisper, peerid, nil).(*LibP2PPeer)
	server.PeerMutex.Lock()
	server.Peers = append(server.Peers, newPeer)
	server.PeerMutex.Unlock()

	return newPeer
}

// NewLibP2PWhisperServer creates a new WhisperServer with
// a libp2p backend.
func NewLibP2PWhisperServer(port uint, whisper *Whisper) (WhisperServer, error) {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 384)
	if err != nil {
		return nil, fmt.Errorf("Error creating libp2p server: %v", err)
	}
	nodeID, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("Error creating libp2p server identity: %v pubkey=%v", err, pub)
	}
	serverAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, fmt.Errorf("Error creating libp2p server address: %v port=%d", err, port)
	}

	ps := pstore.NewPeerstore()
	ps.AddPrivKey(nodeID, priv)
	ps.AddPubKey(nodeID, pub)

	network, err := swarm.NewNetwork(context.Background(), []ma.Multiaddr{serverAddr}, nodeID, ps, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating libp2p network: %v port=%d", err, port)
	}
	h := basichost.New(network)

	server := &LibP2PWhisperServer{
		Host:    h,
		Peers:   []*LibP2PPeer{},
		whisper: whisper,
	}
	return server, nil
}
