//go:build darwin

// Сеть на macOS через netstat (без flow talkers по bps).
package network

import (
	"bufio"
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect на macOS использует netstat (частичная поддержка для кросс-ОС баллов).
func (c *Collector) Collect(ctx context.Context) (any, error) {
	stats := &model.NetworkStats{}

	if out, err := c.exec.Run(ctx, "netstat", "-s", "-p", "tcp"); err == nil {
		stats.ProtocolTalkers = append(stats.ProtocolTalkers, parseNetstatProtocol(out, "TCP")...)
	}
	if out, err := c.exec.Run(ctx, "netstat", "-s", "-p", "udp"); err == nil {
		stats.ProtocolTalkers = append(stats.ProtocolTalkers, parseNetstatProtocol(out, "UDP")...)
	}
	sort.Slice(stats.ProtocolTalkers, func(i, j int) bool {
		return stats.ProtocolTalkers[i].Percent > stats.ProtocolTalkers[j].Percent
	})

	if out, err := c.exec.Run(ctx, "netstat", "-an", "-p", "tcp"); err == nil {
		stats.TCPStates = parseDarwinTCPStates(out)
	}

	if out, err := c.exec.Run(ctx, "netstat", "-anv"); err == nil {
		stats.ListeningSockets = parseDarwinListeners(out)
	}

	return stats, nil
}

func parseNetstatProtocol(out []byte, proto string) []model.ProtocolTalker {
	var bytes uint64
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "bytes") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "bytes" && i > 0 {
					v, _ := strconv.ParseUint(fields[i-1], 10, 64)
					bytes += v
				}
			}
		}
	}
	if bytes == 0 {
		return nil
	}
	return []model.ProtocolTalker{{Protocol: proto, Bytes: bytes, Percent: 100}}
}

func parseDarwinTCPStates(out []byte) []model.TCPStateCount {
	counts := make(map[string]uint32)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 || fields[0] != "tcp4" && fields[0] != "tcp6" {
			continue
		}
		state := fields[len(fields)-1]
		counts[state]++
	}
	result := make([]model.TCPStateCount, 0, len(counts))
	for st, n := range counts {
		result = append(result, model.TCPStateCount{State: st, Count: n})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	return result
}

func parseDarwinListeners(out []byte) []model.ListeningSocket {
	var sockets []model.ListeningSocket
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		if fields[0] != "tcp4" && fields[0] != "tcp6" && fields[0] != "udp4" && fields[0] != "udp6" {
			continue
		}
		if fields[1] != "0" && fields[1] != "LISTEN" {
			continue
		}
		proto := "TCP"
		if strings.HasPrefix(fields[0], "udp") {
			proto = "UDP"
		}
		addr := fields[3]
		port := uint32(0)
		if idx := strings.LastIndex(addr, "."); idx >= 0 {
			p, _ := strconv.ParseUint(addr[idx+1:], 10, 32)
			port = uint32(p)
		}
		sockets = append(sockets, model.ListeningSocket{Protocol: proto, Port: port})
	}
	return sockets
}
