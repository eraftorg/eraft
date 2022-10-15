// Copyright [2022] [WellWood] [wellwood-x@googlegroups.com]

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ekvserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eraft-io/eraft/pkg/core/raft"
	eng "github.com/eraft-io/eraft/pkg/engine"
	"github.com/eraft-io/eraft/pkg/log"
	pb "github.com/eraft-io/eraft/pkg/protocol"
	"net"
	"strings"
	"sync"
	"time"
)

type EkvServer struct {
	mu             sync.RWMutex
	rf             *raft.Raft
	id             int
	gid            int
	applyCh        chan *pb.ApplyMsg
	stopApplyCh    chan interface{}
	lastAppliedIdx int
	logEngine      eng.KvStore
	dbEngine       eng.KvStore
	notifyChan     map[int64]chan *pb.EkvCommandResponse
	stm            map[string]string
	pb.UnimplementedRaftServiceServer
}

func MakeEkvServer(peers map[int]string, id int, addr string) *EkvServer {
	var clientEnds []*raft.RaftClientEnd
	for peerId, peerAddr := range peers {
		newEnd := raft.MakeRaftClientEnd(peerAddr, uint64(peerId))
		clientEnds = append(clientEnds, newEnd)
	}
	newApplyCh := make(chan *pb.ApplyMsg)
	logDbEng := eng.KvStoreFactory("leveldb", fmt.Sprintf("./log/%d", id))
	dataDbEng := eng.KvStoreFactory("leveldb", fmt.Sprintf("./data/%d", id))
	newRaft := raft.MakeRaft(clientEnds, id, logDbEng, newApplyCh, 500, 1500)
	ekvServer := &EkvServer{
		rf:         newRaft,
		applyCh:    newApplyCh,
		id:         id,
		dbEngine:   dataDbEng,
		stm:        make(map[string]string, 1),
		notifyChan: make(map[int64]chan *pb.EkvCommandResponse),
	}
	ekvServer.stopApplyCh = make(chan interface{})
	ekvServer.restoreSnapshot(newRaft.ReadSnapshot())
	go ekvServer.ApplyToSTM(ekvServer.stopApplyCh)
	go ekvServer.RunRedisServer(addr)
	return ekvServer
}

func (s *EkvServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	resp := &pb.RequestVoteResponse{}
	s.rf.HandleRequestVote(req, resp)
	return resp, nil
}

func (s *EkvServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	resp := &pb.AppendEntriesResponse{}
	s.rf.HandleAppendEntries(req, resp)
	return resp, nil
}

func (s *EkvServer) Snapshot(ctx context.Context, req *pb.InstallSnapshotRequest) (*pb.InstallSnapshotResponse, error) {
	resp := &pb.InstallSnapshotResponse{}
	s.rf.HandleInstallSnapshot(req, resp)
	return resp, nil
}

func (s *EkvServer) StopApply() {
	close(s.applyCh)
}

func (s *EkvServer) restoreSnapshot(snapData []byte) {
	if snapData == nil {
		return
	}
}

func (s *EkvServer) takeSnapshot(index int) {

}

func (s *EkvServer) getRespNotifyChan(logIdx int64) chan *pb.EkvCommandResponse {
	if _, ok := s.notifyChan[logIdx]; !ok {
		s.notifyChan[logIdx] = make(chan *pb.EkvCommandResponse, 1)
	}
	return s.notifyChan[logIdx]
}

func (s *EkvServer) RunRedisServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.MainLogger.Error().Msgf(err.Error())
		return
	}
	log.MainLogger.Info().Msgf("ekv redis server success listen on: %s", addr)
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.MainLogger.Error().Msgf(err.Error())
			continue
		}
		go s.HandleBuf(conn)
	}
}

// ParseRedisProtocolBuf parse redis cmd from byte seq
//
// example:
//
//*3\r\n$3\r\nset\r\n$1\r\na\r\n$4\r\ntest
//
// [$3 set $1 a $4 test]
// @return []string     cmd args list
// @return int          cmd args count
func (s *EkvServer) ParseRedisProtocolBuf(reqBuf []byte) ([]string, int) {
	pos := 0
	paramLen := 0
	paramIdx := 0
	log.MainLogger.Info().Msgf("red_buf %s \n", reqBuf)
	for i := 0; i < len(reqBuf); i++ {
		if reqBuf[i] == '\r' && reqBuf[i+1] == '\n' {
			pos = pos + 1
			if pos == 1 {
				{
					for j := range reqBuf[1:i] {
						paramLen *= 10
						paramLen += j - 48
					}
					paramIdx = i + 2
				}
			}
		}
	}
	fmt.Printf("param_idx %d \n", paramIdx)
	paramStr := string(reqBuf[paramIdx:])
	param := strings.Split(paramStr, "\r\n")
	return param, paramLen
}

func (s *EkvServer) HandleBuf(conn net.Conn) {
	defer conn.Close()
	for {
		var buf [10240]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			break
		}
		reqBuf := buf[:n]
		param, _ := s.ParseRedisProtocolBuf(reqBuf)
		fmt.Printf("%v", param)

		req := pb.EkvCommandRequest{}
		if len(param) > 0 {
			if param[1] == "COMMAND" {
				conn.Write([]byte("+OK\r\n"))
				continue
			} else if strings.ToLower(param[1]) == "set" {
				req.OpType = pb.OpType_OP_PUT
				req.Key = param[3]
				req.Val = param[5]
			} else if strings.ToLower(param[1]) == "get" {
				req.OpType = pb.OpType_OP_GET
				req.Key = param[3]
			} else {
				conn.Write([]byte("-Unknow command\r\n"))
				continue
			}
		}
		reqBytes, _ := json.Marshal(&req)
		logIdx, _, isLeader := s.rf.Propose(reqBytes)
		if !isLeader {
			conn.Write([]byte(fmt.Sprintf("-This node is not leader\r\n")))
		}

		s.mu.Lock()
		ch := s.getRespNotifyChan(int64(logIdx))
		s.mu.Unlock()

		select {
		case res := <-ch:
			switch req.OpType {
			case pb.OpType_OP_GET:
				if res.ErrCode == int64(pb.ErrCode_KEY_NOT_EXISTS_ERR) {
					conn.Write([]byte("$-1\r\n"))
				} else {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(res.Val), res.Val)))
				}
			case pb.OpType_OP_PUT:
				conn.Write([]byte(fmt.Sprintf("+OK\r\n")))
			}
		case <-time.After(time.Second * 10):
			conn.Write([]byte("-timeout\r\n"))
		}

		go func() {
			s.mu.Lock()
			delete(s.notifyChan, int64(logIdx))
			s.mu.Unlock()
		}()
	}
}

func (s *EkvServer) ApplyToSTM(done <-chan interface{}) {
	for {
		select {
		case <-done:
			return
		case appliedMsg := <-s.applyCh:
			if appliedMsg.CommandValid {
				req := &pb.EkvCommandRequest{}
				resp := &pb.EkvCommandResponse{}
				log.MainLogger.Error().Msgf("ekv server apply %s\n", string(appliedMsg.Command))
				if err := json.Unmarshal(appliedMsg.Command, req); err != nil {
					resp.ErrCode = int64(pb.ErrCode_UNEXPECT_ERR)
				} else {
					switch req.OpType {
					case pb.OpType_OP_GET:
						valBytes, err := s.dbEngine.Get([]byte(req.Key))
						if err != nil {
							log.MainLogger.Error().Msgf(err.Error())
							resp.ErrCode = int64(pb.ErrCode_KEY_NOT_EXISTS_ERR)
						} else {
							resp.Val = string(valBytes)
							resp.ErrCode = int64(pb.ErrCode_NO_ERR)
						}

					case pb.OpType_OP_PUT:
						if err := s.dbEngine.Put([]byte(req.Key), []byte(req.Val)); err != nil {
							log.MainLogger.Error().Msgf(err.Error())
							resp.ErrCode = int64(pb.ErrCode_WRITE_TO_DB_ERR)
						} else {
							resp.Val = "OK"
							resp.ErrCode = int64(pb.ErrCode_NO_ERR)
						}
					}
				}
				s.lastAppliedIdx = int(appliedMsg.CommandIndex)
				if s.rf.GetLogCount() > 20 {
					s.takeSnapshot(int(appliedMsg.CommandIndex))
				}
				s.mu.Lock()
				ch := s.getRespNotifyChan(appliedMsg.CommandIndex)
				s.mu.Unlock()
				ch <- resp
			}
		}
	}
}
