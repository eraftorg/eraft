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

package rdserver

import (
	"fmt"
	"net"
	"strings"

	"github.com/eraft-io/eraft/pkg/log"
)

func RunRdsServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.MainLogger.Error().Msgf(err.Error())
		return
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.MainLogger.Error().Msgf(err.Error())
			continue
		}
		go HandleBuf(conn)
	}
}

func HandleBuf(conn net.Conn) {
	defer conn.Close()
	for {
		var buf [10240]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			break
		}
		req_buf := buf[:n]
		pos := 0
		param_len := 0
		param_idx := 0
		fmt.Printf("red_buf %s \n", req_buf)
		for i := 0; i < len(req_buf); i++ {
			if req_buf[i] == '\r' && req_buf[i+1] == '\n' {
				pos = pos + 1
				if pos == 1 {
					{
						for j := range req_buf[1:i] {
							param_len *= 10
							param_len += j - 48
						}
						param_idx = i + 2
					}
				}
			}
		}
		fmt.Printf("param_idx %d \n", param_idx)
		params_str := string(req_buf[param_idx:])
		params := strings.Split(params_str, "\r\n")
		fmt.Printf("%v", params)
		if len(params) > 0 {
			if params[1] == "COMMAND" {
				conn.Write([]byte("+OK\r\n"))
			} else if strings.ToLower(params[1]) == "set" {
				// TODO:
			} else if strings.ToLower(params[1]) == "get" {
				// TODO:
			} else {
				conn.Write([]byte("-Unknow command\r\n"))
			}
		}
	}
}
