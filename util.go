package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// LocalIP get local address
func LocalIP(rtype string) (net.IP, error) {
	var dst string
	switch rtype {
	case "A", "IPv4":
		dst = "1.1.1.1:53"
	case "AAAA", "IPv6":
		dst = "[2606:4700:4700::1111]:53"
	default:
		return nil, fmt.Errorf("unsupported rtype: %s", rtype)
	}

	conn, err := net.Dial("udp", dst)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

var resolver = &net.Resolver{}

func initResolver(host string) {
	dnsServers := strings.FieldsFunc(host, func(r rune) bool {
		if r == ',' || r == ' ' || r == ';' {
			return true
		}
		return false
	})

	for i, server := range dnsServers {
		if !strings.Contains(server, ":") {
			dnsServers[i] = server + ":53"
		}
	}

	resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			for _, server := range dnsServers {
				conn, err := d.DialContext(ctx, "udp", server)
				if err == nil {
					return conn, nil
				}
				log.Printf("Resolve %s via %s failed %v", address, server, err)
			}
			return nil, errors.New("no dns resolver avaiable")
		},
	}
}

// Resolve resolves names top ip address
func Resolve(host, rtype string) (string, error) {
	var network string
	switch rtype {
	case "A":
		network = "ip4"
	case "AAAA":
		network = "ip6"
	default:
		panic("bad rtype:" + rtype)
	}

	ip, err := resolver.LookupIP(context.Background(), network, host)
	if err != nil {
		return "", err
	}
	if len(ip) == 0 {
		return "", nil
	}
	return ip[0].String(), nil
}
