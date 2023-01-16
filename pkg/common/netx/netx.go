package netx

import (
	"net"
	"time"
)

func IsReachable(host string, port string, timeout time.Duration) (bool, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if conn == nil {
		return false, nil
	}
	defer conn.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}
