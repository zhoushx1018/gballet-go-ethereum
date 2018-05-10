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
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/p2p"
	peer "github.com/libp2p/go-libp2p-peer"
	set "gopkg.in/fatih/set.v0"
)

// LibP2PPeer implements Peer for libp2p
type Peer struct {
	*PeerBase

	id peer.ID

	server *Server

	connectionStream *LibP2PStream
}

func newLibP2PPeer(s *Server, w *Whisper, pid peer.ID, rw p2p.MsgReadWriter) Peer {
	return &LibP2PPeer{
		&PeerBase{
			host:           w,
			ws:             rw,
			trusted:        false,
			powRequirement: 0.0,
			known:          set.New(),
			quit:           make(chan struct{}),
			bloomFilter:    MakeFullNodeBloom(),
			fullNode:       true,
		},
		pid,
		s,
		nil,
	}
}

// ID returns the id of the peer
func (p *LibP2PPeer) ID() string {
	return p.id.String()
}

func (p *LibP2PPeer) handshake() error {
	err := p.handshakeBase()
	if err != nil {
		return fmt.Errorf("peer [%x] %s", p.ID(), err.Error())
	}
	return nil
}

// start initiates the peer updater, periodically broadcasting the whisper packets
// into the network.
func (p *LibP2PPeer) start() {
	go p.update()
	log.Trace("start", "peer", p.ID())
}

// stop terminates the peer updater, stopping message forwarding to it.
func (p *LibP2PPeer) stop() {
	close(p.quit)
	fmt.Println("stop", "peer", p.ID())
}

// update executes periodic operations on the peer, including message transmission
// and expiration.
func (p *LibP2PPeer) update() {
	// Start the tickers for the updates
	expire := time.NewTicker(expirationCycle)
	transmit := time.NewTicker(transmissionCycle)

	// Loop and transmit until termination is requested
updateLoop:
	for {
		select {
		case <-expire.C:
			p.expire()

		case <-transmit.C:
			if err := p.broadcast(); err != nil {
				break updateLoop
			}

		case <-p.quit:
			break updateLoop
		}
	}

	// Cleanup and remove the peer from the list
	p.server.PeerMutex.Lock()
	for i, it := range p.server.Peers {
		if it.id == p.id {
			p.server.Peers = append(p.server.Peers[:i], p.server.Peers[i+1:]...)
			break
		}
	}
	p.server.PeerMutex.Unlock()
}