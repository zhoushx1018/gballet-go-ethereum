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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	inet "github.com/libp2p/go-libp2p-net"
)

// LibP2PStream is a wrapper used to implement the MsgReadWriter
// interface for libp2p's streams.
type LibP2PStream struct {
	UseDeadline bool
	lp2pStream  inet.Stream
	rlpStream   *rlp.Stream
}

func newLibp2pStream(server *Server, lp2pStream inet.Stream) p2p.MsgReadWriter {
	useDeadline, ok := server.settings.Load(useDeadlineIdx)
	if !ok {
		useDeadline = false
	}
	return &LibP2PStream{
		UseDeadline: useDeadline.(bool),
		lp2pStream:  lp2pStream,
		rlpStream:   rlp.NewStream(bufio.NewReader(lp2pStream), 0),
	}
}

// ReadMsg implements the MsgReadWriter interface to read messages
// from lilbp2p streams.
func (stream *LibP2PStream) ReadMsg() (p2p.Msg, error) {
	if stream.UseDeadline {
		stream.lp2pStream.SetReadDeadline(time.Now().Add(expirationCycle))
	}
	msgcode, err := stream.rlpStream.Uint()
	if err != nil {
		return p2p.Msg{}, fmt.Errorf("can't read message code: %v", err)
	}
	if stream.UseDeadline {
		stream.lp2pStream.SetReadDeadline(time.Now().Add(expirationCycle))
	}
	_, size, err := stream.rlpStream.Kind()
	if err != nil {
		return p2p.Msg{}, fmt.Errorf("can't read message size: %v", err)
	}
	// Only the size of the encoded payload is checked, so theoretically
	// the decrypted message could be much bigger. We would need to add
	// a function to rlp.Stream to find out what is the raw size. Since
	// this is still work in progress, we won't bother with that just yet.
	if size > uint64(MaxMessageSize) {
		return p2p.Msg{}, fmt.Errorf("message too large")
	}
	content, err := stream.rlpStream.Raw()
	if err != nil {
		return p2p.Msg{}, fmt.Errorf("can't read message: %v", err)
	}

	return p2p.Msg{Code: msgcode, Size: uint32(len(content)), Payload: bytes.NewReader(content)}, nil
}

// WriteMsg implements the MsgReadWriter interface to write messages
// to lilbp2p streams.
func (stream *LibP2PStream) WriteMsg(msg p2p.Msg) error {
	if stream.UseDeadline {
		stream.lp2pStream.SetWriteDeadline(time.Now().Add(expirationCycle))
	}

	if err := rlp.Encode(stream.lp2pStream, msg.Code); err != nil {
		return err
	}
	_, err := io.Copy(stream.lp2pStream, msg.Payload)
	return err
}