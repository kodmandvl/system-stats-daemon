// Package model описывает доменные структуры снапшота системы.
package model

// SystemStats — агрегированный снимок всех подсистем мониторинга.
type SystemStats struct {
	LoadAvg     *LoadAvgStats
	CPU         *CPUStats
	DisksLoad   *DisksLoadStats
	Filesystems *FilesystemsStats
	Network     *NetworkStats
}

// LoadAvgStats — средняя загрузка системы (1 / 5 / 15 минут).
type LoadAvgStats struct {
	LoadAvg1  float64
	LoadAvg5  float64
	LoadAvg15 float64
}

// CPUStats — доли времени CPU в процентах.
type CPUStats struct {
	UserMode   float64
	SystemMode float64
	Idle       float64
}

// DisksLoadStats — нагрузка по блочным устройствам.
type DisksLoadStats struct {
	Disks map[string]*DiskLoad
}

// DiskLoad — transfers/s и KB/s (read+write).
type DiskLoad struct {
	Tps float64
	Kbs float64
}

// FilesystemsStats — занятость дискового пространства и inode.
type FilesystemsStats struct {
	FS map[FilesystemKey]*FilesystemStats
}

// FilesystemKey идентифицирует ФС по имени и точке монтирования.
type FilesystemKey struct {
	Name      string
	MountedOn string
}

// FilesystemStats — использовано МБ, %, inode.
type FilesystemStats struct {
	UsedMB            float64
	UsedPercent       float64
	UsedInodes        float64
	UsedInodesPercent float64
}

// NetworkStats — top talkers, сокеты, состояния TCP.
type NetworkStats struct {
	ProtocolTalkers  []ProtocolTalker
	FlowTalkers      []FlowTalker
	ListeningSockets []ListeningSocket
	TCPStates        []TCPStateCount
}

// ProtocolTalker — трафик по протоколу (TCP, UDP, …) за интервал сбора.
type ProtocolTalker struct {
	Protocol string
	Bytes    uint64
	Percent  float64
}

// FlowTalker — пара src/dst с bandwidth (bps после усреднения).
type FlowTalker struct {
	Source      string
	Destination string
	Protocol    string
	Bps         float64
}

// ListeningSocket — слушающий TCP/UDP сокет.
type ListeningSocket struct {
	Command  string
	PID      int32
	User     string
	Protocol string
	Port     uint32
}

// TCPStateCount — число соединений в данном состоянии.
type TCPStateCount struct {
	State string
	Count uint32
}
