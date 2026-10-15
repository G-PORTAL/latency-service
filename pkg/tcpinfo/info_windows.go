package tcpinfo

import (
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

// sioTCPInfo SIO_TCP_INFO = _WSAIORW(IOC_VENDOR, 39), available since Windows 10 1703 / Server 2016
const sioTCPInfo = 0xD8000027

// TCPInfo Mirrors TCP_INFO_v0 from mstcpip.h
type TCPInfo struct {
	State             uint32
	Mss               uint32
	ConnectionTimeMs  uint64
	TimestampsEnabled uint8
	RttUs             uint32
	MinRttUs          uint32
	BytesInFlight     uint32
	Cwnd              uint32
	SndWnd            uint32
	RcvWnd            uint32
	RcvBuf            uint32
	BytesOut          uint64
	BytesIn           uint64
	BytesReordered    uint32
	BytesRetrans      uint32
	FastRetrans       uint32
	DupAcksIn         uint32
	TimeoutEpisodes   uint32
	SynRetrans        uint8
}

// GetRTT Return the smoothed round trip time of *net.TCPConn
func GetRTT(tcpConn *net.TCPConn) (time.Duration, error) {
	info, err := GetsockoptTCPInfo(tcpConn)
	if err != nil {
		return 0, err
	}

	return time.Duration(info.RttUs) * time.Microsecond, nil
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

	version := uint32(0)
	tcpInfo := TCPInfo{}
	var bytesReturned uint32
	var ioctlErr error
	err = rawConn.Control(func(fd uintptr) {
		ioctlErr = syscall.WSAIoctl(syscall.Handle(fd), sioTCPInfo,
			(*byte)(unsafe.Pointer(&version)), uint32(unsafe.Sizeof(version)),
			(*byte)(unsafe.Pointer(&tcpInfo)), uint32(unsafe.Sizeof(tcpInfo)),
			&bytesReturned, nil, 0)
	})
	if err != nil {
		return nil, fmt.Errorf("rawconn control failed. err=%v", err)
	}

	if ioctlErr != nil {
		return nil, fmt.Errorf("WSAIoctl(SIO_TCP_INFO) failed. err=%v", ioctlErr)
	}

	return &tcpInfo, nil
}
