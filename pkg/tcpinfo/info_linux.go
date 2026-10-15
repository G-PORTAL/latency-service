package tcpinfo

import (
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

type TCPInfo syscall.TCPInfo

// GetRTT Return the smoothed round trip time of *net.TCPConn
func GetRTT(tcpConn *net.TCPConn) (time.Duration, error) {
	info, err := GetsockoptTCPInfo(tcpConn)
	if err != nil {
		return 0, err
	}

	return time.Duration(info.Rtt) * time.Microsecond, nil
}

// GetsockoptTCPInfo Return TCPInfo from *net.TCPConn
func GetsockoptTCPInfo(tcpConn *net.TCPConn) (*TCPInfo, error) {
	if tcpConn == nil {
		return nil, fmt.Errorf("tcp conn is nil")
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("error getting raw connection. err=%v", err)
	}

	tcpInfo := TCPInfo{}
	size := unsafe.Sizeof(tcpInfo)
	var errno syscall.Errno
	err = rawConn.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_TCP, syscall.TCP_INFO,
			uintptr(unsafe.Pointer(&tcpInfo)), uintptr(unsafe.Pointer(&size)), 0)
	})
	if err != nil {
		return nil, fmt.Errorf("rawconn control failed. err=%v", err)
	}

	if errno != 0 {
		return nil, fmt.Errorf("syscall failed. errno=%d", errno)
	}

	return &tcpInfo, nil
}
