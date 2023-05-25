target: ~/bin/neo.b

~/bin/neo.b: neo
	mv neo ~/bin/neo.b

clean::
	rm -f ~/bin/neo.b
