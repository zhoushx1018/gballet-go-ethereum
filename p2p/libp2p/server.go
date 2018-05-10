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
	"sync"
	"github.com/ethereum/go-ethereum/p2p"
	host "github.com/libp2p/go-libp2p-host"
)

type Server struct {
    Protocols []p2p.Protocol
	// ...
	
	Host host.Host // Libp2p host structure

	PeerMutex sync.RWMutex  // Guard the list of active peers
	Peers     []*Peer // List of active peers
}

func (s *Server) init() {
    for _, p := range s.Protocols {
        name := fmt.Sprintf("devp2p/%s/%d", p.Name, p.Version)
        server.Host.SetStreamHandler(name, func(stream inet.Stream) {
            // peer := p2p.NewPeer(...)
            go p.Run(peer, &devp2pStreamWrapper{stream})
        })
    }
}