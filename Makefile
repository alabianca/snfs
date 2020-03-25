
install: deps
	go install cli/snfs.go
	go install snfs/main/snfsd_node.go
	go install server/snfsd.go

server: deps
	go build -o bin/snfsd_node ./snfs/main/

cli: deps
	go build -o bin/snfs ./cli/

deps:
	go get -v ./...