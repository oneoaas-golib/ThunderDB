/*
 * Copyright 2018 The ThunderDB Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpc

import (
	"encoding/binary"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"

	"github.com/hashicorp/yamux"
	log "github.com/sirupsen/logrus"
	"github.com/thunderdb/ThunderDB/proto"
)

// ConnType indicating rpc type for multiplexed rpc port
type ConnType byte

// conn type used for incoming connections
const (
	// keep numbers unique.
	// iota depends on order
	ConnTypeRPC  ConnType = 0
	ConnTypeRaft          = 1

	// max raftID length
	MaxRaftIDLength = 256
)

// ServiceMap map service name to service instance
type ServiceMap map[string]interface{}

// Server is the RPC server struct
type Server struct {
	node       *proto.Node
	sessions   sync.Map // map[id]*Session
	rpcServer  *rpc.Server
	stopCh     chan interface{}
	serviceMap ServiceMap
	listener   net.Listener
	raftLayers sync.Map // map[raftID]*RaftLayer
}

// NewServer return a new Server
func NewServer() *Server {
	return &Server{
		rpcServer:  rpc.NewServer(),
		stopCh:     make(chan interface{}),
		serviceMap: make(ServiceMap),
	}
}

// NewServerWithService also return a new Server, and also register the Server.ServiceMap
func NewServerWithService(serviceMap ServiceMap) (server *Server, err error) {
	server = NewServer()
	for k, v := range serviceMap {
		err = server.RegisterService(k, v)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	return server, nil
}

// SetListener set the service loop listener, used by func Serve main loop
func (s *Server) SetListener(l net.Listener) {
	s.listener = l
	return
}

// Serve start the Server main loop,
func (s *Server) Serve() {
serverLoop:
	for {
		select {
		case <-s.stopCh:
			log.Info("Stopping Server Loop")
			break serverLoop
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Debug(err)
				continue
			}
			go s.handleConn(conn)
		}
	}
}

// BindRaftLayer bind raft to server connection
func (s *Server) BindRaftLayer(raftID string, layer *RaftLayer) {
	s.raftLayers.Store(raftID, layer)
}

// UnbindRaftLayer unbind raft to server connection
func (s *Server) UnbindRaftLayer(raftID string) {
	s.raftLayers.Delete(raftID)
}

// handleConn do all the work
func (s *Server) handleConn(conn net.Conn) {
	// Read a single byte
	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if err != io.EOF {
			log.Printf("failed to read byte: %v %s", err, conn.RemoteAddr())
		}
		conn.Close()
		return
	}

	connType := ConnType(buf[0])

	switch connType {
	case ConnTypeRPC:
		s.handleRPC(conn)

	case ConnTypeRaft:
		s.handleRaft(conn)

	default:
		log.Printf("unrecognized RPC byte: %v %s", connType, conn.RemoteAddr())
		conn.Close()
		return
	}
}

// handle ConnTypeRPC connection
func (s *Server) handleRPC(conn net.Conn) {
	defer conn.Close()

	sess, err := yamux.Server(conn, nil)
	if err != nil {
		log.Error(err)
		return
	}

	s.serveRPC(sess)
	log.Debugf("%s closed connection", conn.RemoteAddr())
}

// handle ConnTypeRaft connection
func (s *Server) handleRaft(conn net.Conn) {
	// read database raftID string
	var length uint32
	if err := binary.Read(conn, binary.LittleEndian, &length); err != nil {
		if err != io.EOF {
			log.Warningf("failed to read raftID length: %v %s", err, conn.RemoteAddr())
		}
		conn.Close()
		return
	}

	if length == 0 || length > MaxRaftIDLength {
		log.Warningf("invalid raftID length: %v %s", length, conn.RemoteAddr())
		conn.Close()
		return
	}

	buf := make([]byte, length)

	if _, err := io.ReadFull(conn, buf); err != nil {
		if err != io.EOF {
			log.Warningf("failed to read raftID length: %v %s", err, conn.RemoteAddr())
		}
		conn.Close()
		return
	}

	raftID := string(buf)

	if layer, ok := s.raftLayers.Load(raftID); ok {
		layer.(*RaftLayer).Handoff(conn)
	} else {
		log.Warningf("failed to found raft layer for raftID: %v %s", raftID, conn.RemoteAddr())
		conn.Close()
	}
}

// serveRPC install the JSON RPC codec
func (s *Server) serveRPC(sess *yamux.Session) {
	conn, err := sess.Accept()
	if err != nil {
		log.Error(err)
		return
	}
	s.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
}

// RegisterService with a Service name, used by Client RPC
func (s *Server) RegisterService(name string, service interface{}) error {
	return s.rpcServer.RegisterName(name, service)
}

// Stop Server main loop
func (s *Server) Stop() {
	close(s.stopCh)
}
