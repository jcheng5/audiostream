all: bin/server bin/client

bin/server: server/*.go
	go build -o bin/server ./server

bin/client: client/*.go
	go build -o bin/client ./client
