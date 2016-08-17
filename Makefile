all: bin/source bin/sink

bin/source: source/*.go
	go build -o bin/source ./source

bin/sink: sink/*.go
	go build -o bin/sink ./sink
