
install: deps
	go install cli/snfs.go
	go install snfs/main/snfsd.go

server: deps
	go build -o bin/snfsd ./snfs/main/

cli: deps
	go build -o bin/snfs ./cli/

deps:
	go get -v ./...