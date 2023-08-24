//go:build linux
// +build linux

package main

import (
	"github.com/g-portal/latency-service/pkg/server"
)

func main() {
	server.StartServer()
}
