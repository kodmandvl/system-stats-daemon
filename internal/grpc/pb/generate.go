// Package pb содержит сгенерированные из api/stats.proto типы и gRPC-код.
package pb

//go:generate protoc --go_out=../../../ --go_opt=module=github.com/kodmandvl/system-stats-daemon --go_opt=paths=import --go-grpc_out=../../../ --go-grpc_opt=module=github.com/kodmandvl/system-stats-daemon --go-grpc_opt=paths=import -I../../../api ../../../api/stats.proto
