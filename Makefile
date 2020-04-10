
install: deps
	go install ./cmd/snfs/main.go
	go install ./cmd/snfsd/main.go
	go install ./cmd/snfsd_node/main.go

snfs_node: deps
	go build -o bin/snfsd_node ./cmd/snfs_node

snfsd: deps
	go build -o bin/snfsd ./cmd/snfsd/

cli: deps
	go build -o bin/snfs ./cmd/snfs/

deps:
	go get -v ./...