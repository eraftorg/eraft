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

package main

import (
	"flag"
	"fmt"
	"github.com/eraft-io/eraft/pkg/common"
	"github.com/eraft-io/eraft/pkg/consts"
	"github.com/eraft-io/eraft/pkg/ekvserver"
	"github.com/eraft-io/eraft/pkg/log"
	pb "github.com/eraft-io/eraft/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var ekvServerId = flag.Int("id", 0, "ekv server id")
var ekvServerPeers = flag.String("peers", "127.0.0.1:7088,127.0.0.1:7089,127.0.0.1:7090", "ekv server peers(eg: 127.0.0.1:7088,127.0.0.1:7089,127.0.0.1:7090)")
var ekvServerAddress = flag.String("address", "127.0.0.1:17088,127.0.0.1:17089,127.0.0.1:17090", "ekv server addresses(eg: 127.0.0.1:17088,127.0.0.1:17089,127.0.0.1:17090)")

func main() {
	flag.Parse()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ekvServerPeersMap := common.ConfigStringToMap(*ekvServerPeers)
	ekvServerAddressMap := common.ConfigStringToMap(*ekvServerAddress)
	ekvServer := ekvserver.MakeEkvServer(ekvServerPeersMap, *ekvServerId, ekvServerAddressMap[*ekvServerId])
	svr := grpc.NewServer(grpc.MaxSendMsgSize(consts.MAX_GRPC_SEND_MSG_SIZE), grpc.MaxRecvMsgSize(consts.MAX_GRPC_RECV_MSG_SIZE))
	pb.RegisterRaftServiceServer(svr, ekvServer)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)
	go func() {
		sig := <-sigs
		fmt.Printf("recived sig %d \n", sig)
		ekvServer.StopApply()
		os.Exit(-1)
	}()
	reflection.Register(svr)
	lis, err := net.Listen("tcp", ekvServerPeersMap[*ekvServerId])
	if err != nil {
		log.MainLogger.Error().Msgf("ekv server failed to listen: %v", err)
		return
	}
	log.MainLogger.Info().Msgf("ekv server success listen on: %s", ekvServerPeersMap[*ekvServerId])
	if err := svr.Serve(lis); err != nil {
		log.MainLogger.Error().Msgf("ekv server failed to serve: %v", err)
		return
	}
}
