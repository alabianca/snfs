server: deps
	go build -o bin/snfsd ./snfs/main/

cli: deps
	go build -o bin/snfs ./cli/

deps:
	go get -v ./...