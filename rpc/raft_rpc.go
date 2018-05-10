/*
 * Copyright (c) 2013-2018. HashiCorp, Inc.
 * Copyright 2018 The ThunderDB Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the “License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an “AS IS” BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpc

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rqlite/rqlite/store"
)

// assert interface coerce
var _ store.Listener = &RaftLayer{}

// RaftLayer implements the raft.StreamLayer interface,
// so that we can use a single RPC layer for Raft and Consul
type RaftLayer struct {
	// raftID is the raft instance id for raft rpc port multiplexing
	raftID string

	// src is the address for outgoing connections.
	src net.Addr

	// addr is the listener address to return.
	addr net.Addr

	// connCh is used to accept connections.
	connCh chan net.Conn

	// Tracks if we are closed
	closed    bool
	closeCh   chan struct{}
	closeLock sync.Mutex
}

// NewRaftLayer is used to initialize a new RaftLayer which can
// be used as a StreamLayer for Raft.
func NewRaftLayer(src, addr net.Addr, raftID string) *RaftLayer {
	layer := &RaftLayer{
		raftID:  raftID,
		src:     src,
		addr:    addr,
		connCh:  make(chan net.Conn),
		closeCh: make(chan struct{}),
	}
	return layer
}

// Handoff is used to hand off a connection to the
// RaftLayer. This allows it to be Accept()'ed
func (l *RaftLayer) Handoff(c net.Conn) error {
	select {
	case l.connCh <- c:
		return nil
	case <-l.closeCh:
		return fmt.Errorf("Raft RPC layer closed")
	}
}

// Accept is used to return connection which are
// dialed to be used with the Raft layer
func (l *RaftLayer) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closeCh:
		return nil, fmt.Errorf("Raft RPC layer closed")
	}
}

// Close is used to stop listening for Raft connections
func (l *RaftLayer) Close() error {
	l.closeLock.Lock()
	defer l.closeLock.Unlock()

	if !l.closed {
		l.closed = true
		close(l.closeCh)
	}
	return nil
}

// Addr is used to return the address of the listener
func (l *RaftLayer) Addr() net.Addr {
	return l.addr
}

// Dial is used to create a new outgoing connection
func (l *RaftLayer) Dial(address string, timeout time.Duration) (net.Conn, error) {
	d := &net.Dialer{LocalAddr: l.src, Timeout: timeout}
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// Write the Raft byte to set the mode
	_, err = conn.Write([]byte{byte(ConnTypeRaft)})
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Write the raftID info to set the connection raft instance
	var raftIDLen uint32
	raftIDLen = uint32(len(l.raftID))
	err = binary.Write(conn, binary.LittleEndian, raftIDLen)
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = conn.Write([]byte(l.raftID))
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, err
}
