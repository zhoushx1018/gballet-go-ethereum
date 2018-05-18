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

package libp2p

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/discover"

	"github.com/ethereum/go-ethereum/p2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	swarm "github.com/libp2p/go-libp2p-swarm"

	// crypto "github.com/libp2p/go-libp2p-crypto"
	// inet "github.com/libp2p/go-libp2p-net"
	// peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	// set "gopkg.in/fatih/set.v0"
)

type Server struct {
	Protocols []p2p.Protocol
	NodeID    discover.NodeID

	Host host.Host // Libp2p host structure

	useDeadlines bool

	// PeerMutex sync.RWMutex // Guard the list of active peers
	// Peers     []*p2p.Peer  // List of active peers
}

func NewServer(protocols []p2p.Protocol, port uint) (*Server, error) {
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

	server := &Server{
		Host:      h,
		Protocols: protocols,
	}

	return server, nil
}

func (server *Server) Init() {
	for _, p := range server.Protocols {
		name := fmt.Sprintf("/devp2p/%s/%d", p.Name, p.Version)
		server.Host.SetStreamHandler(protocol.ID(name), func(stream inet.Stream) {
			// In the case of Whisper there is an issue that will not
			peer := &p2p.Peer{}
			go p.Run(peer, newStream(server.useDeadlines, stream))
		})
	}
}
