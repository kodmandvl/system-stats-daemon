//go:build linux

// Сеть на Linux: /proc/net/snmp, ss.
package network

import (
	"bufio"
	"context"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// commandRunner — минимальный интерфейс для ss/netstat (совместим с collector.CommandExecutor).
type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// prevSnmp хранит предыдущие счётчики для расчёта дельт за интервал сбора.
var (
	prevSnmp   map[string]uint64
	prevSnmpMu sync.Mutex
)

// Collect параллельно не вызывается — ошибки отдельных источников игнорируются.
func (c *Collector) Collect(ctx context.Context) (any, error) {
	stats := &model.NetworkStats{}

	protocolTalkers, err := collectProtocolTalkers()
	if err == nil {
		stats.ProtocolTalkers = protocolTalkers
	}

	tcpStates, err := collectTCPStates(ctx, c.exec)
	if err == nil {
		stats.TCPStates = tcpStates
	}

	listeners, err := collectListeners(ctx, c.exec)
	if err == nil {
		stats.ListeningSockets = listeners
	}

	flows, err := collectFlowTalkers(ctx, c.exec)
	if err == nil {
		stats.FlowTalkers = flows
	}

	return stats, nil
}

func collectProtocolTalkers() ([]model.ProtocolTalker, error) {
	data, err := os.ReadFile("/proc/net/snmp")
	if err != nil {
		return nil, err
	}

	curr := parseSnmpCounters(data)
	prevSnmpMu.Lock()
	defer prevSnmpMu.Unlock()

	if prevSnmp == nil {
		prevSnmp = curr
		return nil, nil
	}

	var talkers []model.ProtocolTalker
	var total uint64
	for proto, keys := range map[string][]string{
		"TCP":  {"InSegs", "OutSegs"},
		"UDP":  {"InDatagrams", "OutDatagrams"},
		"ICMP": {"InMsgs", "OutMsgs"},
		"IP":   {"InOctets", "OutOctets"},
	} {
		var delta uint64
		for _, k := range keys {
			delta += curr[proto+":"+k] - prevSnmp[proto+":"+k]
		}
		if delta > 0 {
			talkers = append(talkers, model.ProtocolTalker{Protocol: proto, Bytes: delta})
			total += delta
		}
	}
	prevSnmp = curr

	for i := range talkers {
		if total > 0 {
			talkers[i].Percent = float64(talkers[i].Bytes) / float64(total) * 100
		}
	}
	sort.Slice(talkers, func(i, j int) bool {
		return talkers[i].Percent > talkers[j].Percent
	})
	return talkers, nil
}

func parseSnmpCounters(data []byte) map[string]uint64 {
	result := make(map[string]uint64)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var headers []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Ip:") || strings.HasPrefix(line, "Icmp:") ||
			strings.HasPrefix(line, "Tcp:") || strings.HasPrefix(line, "Udp:") {
			parts := strings.Fields(line)
			proto := strings.TrimSuffix(parts[0], ":")
			if strings.HasSuffix(line, "InReceives") || strings.Contains(line, "InSegs") ||
				len(parts) > 1 && !isNumber(parts[1]) {
				headers = parts[1:]
				continue
			}
			if len(headers) == 0 {
				continue
			}
			for i, val := range parts[1:] {
				if i >= len(headers) {
					break
				}
				n, _ := strconv.ParseUint(val, 10, 64)
				result[proto+":"+headers[i]] = n
			}
		}
	}
	return result
}

func isNumber(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func collectTCPStates(ctx context.Context, exec commandRunner) ([]model.TCPStateCount, error) {
	out, err := exec.Run(ctx, "ss", "-ta")
	if err != nil {
		return collectTCPStatesFromProc()
	}

	counts := make(map[string]uint32)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 1 {
			continue
		}
		state := fields[0]
		if state == "State" || state == "Netid" {
			continue
		}
		if strings.HasPrefix(state, "LISTEN") || strings.Contains(state, ":") {
			continue
		}
		counts[state]++
	}

	return mapToTCPStates(counts), nil
}

func collectTCPStatesFromProc() ([]model.TCPStateCount, error) {
	data, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		return nil, err
	}
	states := map[string]string{
		"01": "ESTAB", "02": "SYN_SENT", "03": "SYN_RECV", "04": "FIN_WAIT1",
		"05": "FIN_WAIT2", "06": "TIME_WAIT", "07": "CLOSE", "08": "CLOSE_WAIT",
		"09": "LAST_ACK", "0A": "LISTEN", "0B": "CLOSING",
	}
	counts := make(map[string]uint32)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo == 1 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		st := strings.ToUpper(fields[3])
		if name, ok := states[st]; ok {
			counts[name]++
		}
	}
	return mapToTCPStates(counts), nil
}

func mapToTCPStates(counts map[string]uint32) []model.TCPStateCount {
	result := make([]model.TCPStateCount, 0, len(counts))
	for state, cnt := range counts {
		result = append(result, model.TCPStateCount{State: state, Count: cnt})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	return result
}

func collectListeners(ctx context.Context, exec commandRunner) ([]model.ListeningSocket, error) {
	out, err := exec.Run(ctx, "ss", "-lntup")
	if err != nil {
		return nil, err
	}

	var sockets []model.ListeningSocket
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Netid") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		proto := strings.ToUpper(fields[0])
		addr := fields[4]
		port := parsePort(addr)
		pid, cmd := parseProcessInfo(line)

		sockets = append(sockets, model.ListeningSocket{
			Command:  cmd,
			PID:      pid,
			User:     "", // ss -lntup не всегда отдаёт user в парсируемом виде
			Protocol: proto,
			Port:     port,
		})
	}
	return sockets, nil
}

func parsePort(addr string) uint32 {
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		p, _ := strconv.ParseUint(addr[idx+1:], 10, 32)
		return uint32(p)
	}
	return 0
}

// parseProcessInfo извлекает pid и имя процесса из строки ss (users:(("sshd",pid=123,...))).
func parseProcessInfo(line string) (int32, string) {
	idx := strings.Index(line, "users:")
	if idx < 0 {
		return 0, ""
	}
	chunk := line[idx:]
	return parsePIDFromUsers(chunk), parseCommandFromUsers(chunk)
}

func parsePIDFromUsers(chunk string) int32 {
	p := strings.Index(chunk, "pid=")
	if p < 0 {
		return 0
	}
	end := strings.IndexAny(chunk[p+4:], ",)")
	if end < 0 {
		end = len(chunk[p+4:])
	}
	v, _ := strconv.ParseInt(chunk[p+4:p+4+end], 10, 32)
	return int32(v)
}

func parseCommandFromUsers(chunk string) string {
	q1 := strings.Index(chunk, `("`)
	if q1 < 0 {
		return ""
	}
	q2 := strings.Index(chunk[q1+2:], `"`)
	if q2 <= 0 {
		return ""
	}
	return chunk[q1+2 : q1+2+q2]
}

type ssFlowKey struct {
	src, dst, proto string
}

func collectFlowTalkers(ctx context.Context, exec commandRunner) ([]model.FlowTalker, error) {
	out, err := exec.Run(ctx, "ss", "-ti")
	if err != nil {
		return nil, err
	}

	type acc struct {
		bytes uint64
	}
	flows := make(map[ssFlowKey]*acc)

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var current ssFlowKey
	var hasCurrent bool
	for scanner.Scan() {
		line := scanner.Text()
		if key, ok := parseSSFlowLine(line); ok {
			current = key
			hasCurrent = true
			if flows[current] == nil {
				flows[current] = &acc{}
			}
			continue
		}
		if !hasCurrent {
			continue
		}
		for _, part := range strings.Fields(line) {
			if strings.HasPrefix(part, "bytes_acked:") {
				v, _ := strconv.ParseUint(strings.TrimPrefix(part, "bytes_acked:"), 10, 64)
				if flows[current] != nil {
					flows[current].bytes += v
				}
			}
		}
	}

	talkers := make([]model.FlowTalker, 0, len(flows))
	for k, a := range flows {
		if a.bytes == 0 {
			continue
		}
		talkers = append(talkers, model.FlowTalker{
			Source:      k.src,
			Destination: k.dst,
			Protocol:    k.proto,
			Bps:         float64(a.bytes), // усреднение выполнит storage
		})
	}
	sort.Slice(talkers, func(i, j int) bool { return talkers[i].Bps > talkers[j].Bps })
	if len(talkers) > 20 {
		talkers = talkers[:20]
	}
	return talkers, nil
}

// parseSSFlowLine разбирает строку соединения из ss -ti (формат с netid и без).
func parseSSFlowLine(line string) (ssFlowKey, bool) {
	if strings.HasPrefix(line, "\t") || strings.Contains(line, "bytes_acked:") {
		return ssFlowKey{}, false
	}
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return ssFlowKey{}, false
	}
	if fields[0] == "State" || fields[0] == "Netid" {
		return ssFlowKey{}, false
	}

	var proto, local, peer string
	switch {
	case fields[0] == "tcp" || fields[0] == "udp":
		if len(fields) < 6 {
			return ssFlowKey{}, false
		}
		proto = strings.ToUpper(fields[0])
		local, peer = fields[4], fields[5]
	default:
		// ESTAB 0 0 127.0.0.1:port 127.0.0.1:port
		local, peer = fields[len(fields)-2], fields[len(fields)-1]
		proto = "TCP"
	}

	if !strings.Contains(local, ":") || !strings.Contains(peer, ":") {
		return ssFlowKey{}, false
	}

	return ssFlowKey{proto: proto, src: local, dst: peer}, true
}
