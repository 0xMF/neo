target:~/bin/neo

~/bin/neo: src/*.go
	go fmt src/neo.go src/types.go src/ui.go src/local.go
	go build -o neo src/neo.go src/types.go src/ui.go src/local.go
	chmod 711 neo
	mv neo ~/bin/neo
