//go:build linux
// +build linux

package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/g-portal/latency-service/pkg/logging"
	"github.com/g-portal/latency-service/pkg/tcpinfo"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type PingResponse struct {
	SourceIP  string    `json:"source_ip"`
	LatencyMS int64     `json:"latency_ms"`
	Time      time.Time `json:"time"`
}

type contextKey struct {
	key string
}

var ConnContextKey = &contextKey{"http-conn"}

func saveConnIntoContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, ConnContextKey, c)
}

// getConn Get *net.TCPConn depending if TLS or TCP connection
func getConn(r *http.Request) *net.TCPConn {
	if netConn, ok := r.Context().Value(ConnContextKey).(net.Conn); ok {
		// Return *net.TCPConn if HTTP
		if tcpConn, ok := netConn.(*net.TCPConn); ok {
			return tcpConn
		}

		// Return *net.TCPConn inside *tls.Conn when HTTPS
		// *tls.Conn contains the *net.TCPConn since Golang 1.18
		if tlsConn, ok := netConn.(*tls.Conn); ok {
			if tcpConn, ok := tlsConn.NetConn().(*net.TCPConn); ok {
				return tcpConn
			}
		}
	}

	return nil
}

// ping Ping Endpoint
func ping() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
                        w.Header().Set("Access-Control-Allow-Headers", "*")
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tcpConnection := getConn(r)
		if tcpConnection == nil {
			log.Printf("Failed to receive TCP connection")
			handleError(w)
			return
		}

		info, err := tcpinfo.GetsockoptTCPInfo(tcpConnection)
		if err != nil {
			log.Printf("Failed to get socket information: %s", err)
			handleError(w)
			return
		}

		respBytes, err := json.Marshal(PingResponse{
			SourceIP:  strings.Split(r.RemoteAddr, ":")[0],
			LatencyMS: time.Duration(int64(info.Rtt * 1000)).Milliseconds(),
			Time:      time.Now(),
		})

		if err != nil {
			log.Printf("Failed to marshal json: %s", err)
			handleError(w)
			return
		}

		logging.LogToFile(respBytes)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Headers", "*")
		_, _ = w.Write(respBytes)
	})
}
