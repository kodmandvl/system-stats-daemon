// Преобразование domain → protobuf для gRPC.
package server

import (
	"github.com/kodmandvl/system-stats-daemon/internal/grpc/pb"
	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// toProto маппит внутреннюю модель в сообщения api/stats.proto.
func toProto(s *model.SystemStats) *pb.SystemStats {
	if s == nil {
		return &pb.SystemStats{}
	}
	msg := &pb.SystemStats{}
	if s.LoadAvg != nil {
		msg.LoadAvg = &pb.LoadAvgStats{
			LoadAvg1: s.LoadAvg.LoadAvg1, LoadAvg5: s.LoadAvg.LoadAvg5, LoadAvg15: s.LoadAvg.LoadAvg15,
		}
	}
	if s.CPU != nil {
		msg.Cpu = &pb.CPUStats{
			UserMode: s.CPU.UserMode, SystemMode: s.CPU.SystemMode, Idle: s.CPU.Idle,
		}
	}
	if s.DisksLoad != nil && len(s.DisksLoad.Disks) > 0 {
		msg.DisksLoad = &pb.DisksLoadStats{}
		for name, d := range s.DisksLoad.Disks {
			msg.DisksLoad.Disks = append(msg.DisksLoad.Disks, &pb.DiskLoad{
				Device: name, Tps: d.Tps, Kbs: d.Kbs,
			})
		}
	}
	if s.Filesystems != nil {
		msg.Filesystems = &pb.FilesystemsStats{}
		for k, fs := range s.Filesystems.FS {
			msg.Filesystems.Filesystems = append(msg.Filesystems.Filesystems, &pb.FilesystemStats{
				Filesystem: k.Name, MountedOn: k.MountedOn,
				UsedMb: fs.UsedMB, UsedPercent: fs.UsedPercent,
				UsedInodes: fs.UsedInodes, UsedInodesPercent: fs.UsedInodesPercent,
			})
		}
	}
	if s.Network != nil {
		msg.Network = &pb.NetworkStats{}
		for _, t := range s.Network.ProtocolTalkers {
			msg.Network.ProtocolTalkers = append(msg.Network.ProtocolTalkers, &pb.ProtocolTalker{
				Protocol: t.Protocol, Bytes: t.Bytes, Percent: t.Percent,
			})
		}
		for _, f := range s.Network.FlowTalkers {
			msg.Network.FlowTalkers = append(msg.Network.FlowTalkers, &pb.FlowTalker{
				Source: f.Source, Destination: f.Destination, Protocol: f.Protocol, Bps: f.Bps,
			})
		}
		for _, l := range s.Network.ListeningSockets {
			msg.Network.ListeningSockets = append(msg.Network.ListeningSockets, &pb.ListeningSocket{
				Command: l.Command, Pid: l.PID, User: l.User, Protocol: l.Protocol, Port: l.Port,
			})
		}
		for _, st := range s.Network.TCPStates {
			msg.Network.TcpStates = append(msg.Network.TcpStates, &pb.TCPStateCount{
				State: st.State, Count: st.Count,
			})
		}
	}
	return msg
}
