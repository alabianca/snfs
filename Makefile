
server: deps
	go build -o bin/snfsd ./snfs/main/

deps:
	go get -v ./...
