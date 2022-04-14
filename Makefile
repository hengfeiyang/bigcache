build:
	go build -o bigcache cmd/main.go

run: build
	GODEBUG=gctrace=1 ./bigcache

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem -benchtime=4s . -timeout 30m  -cpuprofile=cpu.pprof -memprofile=mem.pprof
