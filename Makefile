target:~/bin/neo

~/bin/neo: src/*.go
	go build -o neo src/neo.go src/ui.go src/local.go
	mv -i neo ~/bin/neo
