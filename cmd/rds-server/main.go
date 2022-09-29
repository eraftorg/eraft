package main

import "github.com/eraft-io/eraft/pkg/rdserver"

func main() {
	rdserver.RunRdsServer("127.0.0.1:12306")
}
