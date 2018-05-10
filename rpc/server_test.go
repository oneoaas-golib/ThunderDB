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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thunderdb/ThunderDB/utils"
)

type TestService struct {
	counter int
}

type TestReq struct {
	Step int
}

type TestRep struct {
	Ret int
}

func NewTestService() *TestService {
	return &TestService{
		counter: 0,
	}
}

func (s *TestService) IncCounter(req *TestReq, rep *TestRep) error {
	log.Debugf("calling IncCounter req:%v, rep:%v", *req, *rep)
	s.counter += req.Step
	rep.Ret = s.counter
	return nil
}

func (s *TestService) IncCounterSimpleArgs(step int, ret *int) error {
	log.Debugf("calling IncCounter req:%v, rep:%v", step, ret)
	s.counter += step
	*ret = s.counter
	return nil
}

func TestIncCounter(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server, err := NewServerWithService(ServiceMap{"Test": NewTestService()})
	server.SetListener(l)
	go server.Serve()

	rep := new(TestRep)
	client, err := InitClient(l.Addr().String())
	if err != nil {
		log.Fatal(err)
	}

	err = client.Call("Test.IncCounter", &TestReq{Step: 10}, rep)
	if err != nil {
		log.Fatal(err)
	}
	utils.CheckNum(rep.Ret, 10, t)

	err = client.Call("Test.IncCounter", &TestReq{Step: 10}, rep)
	if err != nil {
		log.Fatal(err)
	}
	utils.CheckNum(rep.Ret, 20, t)

	repSimple := new(int)
	err = client.Call("Test.IncCounterSimpleArgs", 10, repSimple)
	if err != nil {
		log.Fatal(err)
	}
	utils.CheckNum(*repSimple, 30, t)

	client.Close()
	server.Stop()
}

func TestIncCounterSimpleArgs(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server, err := NewServerWithService(ServiceMap{"Test": NewTestService()})
	server.SetListener(l)
	go server.Serve()

	client, err := InitClient(l.Addr().String())
	if err != nil {
		log.Fatal(err)
	}

	repSimple := new(int)
	err = client.Call("Test.IncCounterSimpleArgs", 10, repSimple)
	if err != nil {
		log.Fatal(err)
	}
	utils.CheckNum(*repSimple, 10, t)

	client.Close()
	server.Stop()
}

func TestServer_Close(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()
	testService := NewTestService()
	err = server.RegisterService("Test", testService)
	if err != nil {
		log.Fatal(err)
	}
	server.SetListener(l)
	go server.Serve()

	server.Stop()
}

func TestServerHandoffRaftLayer(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := NewServer()
	server.SetListener(l)
	go server.Serve()

	// build a test raft connection
	lsrc, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	lsrcAddr := lsrc.Addr()
	lsrc.Close()

	// generate random raftID
	raftID := fmt.Sprintf("%v", uuid.Must(uuid.NewV4()))
	r := NewRaftLayer(lsrcAddr, l.Addr(), raftID)
	server.BindRaftLayer(raftID, r)

	// dial itself
	initConn, err := r.Dial(l.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	acceptConn, err := r.Accept()
	if err != nil {
		t.Fatal(err.Error())
	}

	lineData := []byte("happy\n")
	_, err = initConn.Write(lineData)
	acceptLineData, err := bufio.NewReader(acceptConn).ReadSlice('\n')
	if err != nil {
		t.Fatal(err.Error())
	}

	if bytes.Compare(lineData, acceptLineData) != 0 {
		t.Errorf("%s = %s, want %s", string(lineData), string(acceptLineData),
			string(lineData))
	}

	r.Close()
	initConn.Close()
	acceptConn.Close()
	server.Stop()
}

func TestServerHandoffMultipleRaftLayer(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := NewServer()
	server.SetListener(l)
	go server.Serve()

	// build a test raft connection
	l1, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	l1Addr := l1.Addr()
	l1.Close()

	l2, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	l2Addr := l2.Addr()
	l2.Close()

	// generate random raftID
	raftID1 := fmt.Sprintf("%v", uuid.Must(uuid.NewV4()))
	r1 := NewRaftLayer(l1Addr, l.Addr(), raftID1)
	server.BindRaftLayer(raftID1, r1)

	raftID2 := fmt.Sprintf("%v", uuid.Must(uuid.NewV4()))
	r2 := NewRaftLayer(l2Addr, l.Addr(), raftID2)
	server.BindRaftLayer(raftID2, r2)

	// dial itself for r1
	initConn1, err := r1.Dial(l.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	acceptConn1, err := r1.Accept()
	if err != nil {
		t.Fatal(err.Error())
	}

	lineData1 := []byte("happy1\n")
	_, err = initConn1.Write(lineData1)

	// dial itself for r2
	initConn2, err := r2.Dial(l.Addr().String(), time.Second)
	if err != nil {
		t.Fatalf(err.Error())
	}

	acceptConn2, err := r2.Accept()
	if err != nil {
		t.Fatal(err.Error())
	}

	lineData2 := []byte("happy2\n")
	_, err = initConn2.Write(lineData2)

	acceptLineData1, err := bufio.NewReader(acceptConn1).ReadSlice('\n')
	if err != nil {
		t.Fatal(err.Error())
	}
	acceptLineData2, err := bufio.NewReader(acceptConn2).ReadSlice('\n')
	if err != nil {
		t.Fatal(err.Error())
	}

	if bytes.Compare(lineData1, acceptLineData1) != 0 {
		t.Errorf("%s = %s, want %s", string(lineData1), string(acceptLineData1),
			string(lineData1))
	}
	if bytes.Compare(lineData2, acceptLineData2) != 0 {
		t.Errorf("%s = %s, want %s", string(lineData2), string(acceptLineData2),
			string(lineData2))
	}

	r1.Close()
	r2.Close()
	initConn1.Close()
	initConn2.Close()
	acceptConn1.Close()
	acceptConn2.Close()

	server.Stop()
}

func TestServerUnbindHandoffRaftLayer(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := NewServer()
	server.SetListener(l)
	go server.Serve()

	// build a test raft connection
	srcListener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	srcAddr := srcListener.Addr()
	srcListener.Close()

	// generate random raftID
	raftID := fmt.Sprintf("%v", uuid.Must(uuid.NewV4()))
	r := NewRaftLayer(srcAddr, l.Addr(), raftID)

	// bind and unbind
	server.BindRaftLayer(raftID, r)
	server.UnbindRaftLayer(raftID)

	// dial itself
	initConn, err := r.Dial(l.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	complete := make(chan struct{})

	go func() {
		_, err := r.Accept()
		if !strings.Contains(err.Error(), "closed") {
			t.Error(err.Error())
		}
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		t.Fatal("unbind raft layer cannot be accepted")
	case <-time.After(time.Second * 2):
	}

	r.Close()
	initConn.Close()
	server.Stop()
}

func TestServerInvalidConnType(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := NewServer()
	server.SetListener(l)
	go server.Serve()

	// initiate connection with no interaction
	Convey("connect with no interaction", t, func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		So(err, ShouldBeNil)
		conn.Close()
	})

	// initiate connection with invalid ConnType
	Convey("invalid ConnType", t, func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		So(err, ShouldBeNil)
		_, err = conn.Write([]byte{byte(127)})
		So(err, ShouldBeNil)
		conn.SetDeadline(time.Now().Add(time.Second))
		buffer := make([]byte, 1)
		_, err = conn.Read(buffer)
		So(err, ShouldEqual, io.EOF)
		conn.Close()
	})

	server.Stop()
}

func TestServerInvalidRaftConnection(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	addr := "127.0.0.1:0"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err.Error())
	}

	server := NewServer()
	server.SetListener(l)
	go server.Serve()

	// initiate connection with no raft id length
	Convey("no raft id length", t, func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		So(err, ShouldBeNil)
		_, err = conn.Write([]byte{byte(ConnTypeRaft)})
		So(err, ShouldBeNil)
		conn.Close()
	})

	// initiate connection with zero raft id length
	Convey("zero raft id length", t, func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		So(err, ShouldBeNil)
		_, err = conn.Write([]byte{byte(ConnTypeRaft)})
		So(err, ShouldBeNil)
		length := uint32(0)
		err = binary.Write(conn, binary.LittleEndian, length)
		So(err, ShouldBeNil)
		conn.SetDeadline(time.Now().Add(time.Second))
		buffer := make([]byte, 1)
		_, err = conn.Read(buffer)
		So(err, ShouldEqual, io.EOF)
		conn.Close()
	})

	// initiate connection with large raft id length
	Convey("too large raft id length", t, func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		So(err, ShouldBeNil)
		_, err = conn.Write([]byte{byte(ConnTypeRaft)})
		So(err, ShouldBeNil)
		length := uint32(1024)
		err = binary.Write(conn, binary.LittleEndian, length)
		So(err, ShouldBeNil)
		conn.SetDeadline(time.Now().Add(time.Second))
		buffer := make([]byte, 1)
		_, err = conn.Read(buffer)
		So(err, ShouldEqual, io.EOF)
		conn.Close()
	})

	server.Stop()
}
