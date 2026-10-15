//go:build !linux && !windows

package tcpinfo

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

// GetRTT Not supported on this platform
func GetRTT(tcpConn *net.TCPConn) (time.Duration, error) {
	return 0, fmt.Errorf("tcp info is not supported on %s", runtime.GOOS)
}
