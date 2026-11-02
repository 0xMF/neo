include src/local.mk

target:neo

neo: src/*.go
	go fmt src/neo.go src/types.go src/ui.go src/local.go
	CGO_ENABLED=0 go build -o neo -ldflags="-s -w" -trimpath src/neo.go src/types.go src/ui.go src/local.go
	upx neo       # further reduce size with upx
	chmod 700 neo

clean::
	rm -f neo
