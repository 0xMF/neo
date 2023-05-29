EXT :=$(shell git branch --show-current 2> /dev/null | cut -c1)

ifndef $(EXT)
ifeq ($(EXT),)
PWD := $(notdir $(CURDIR))
EXT := $(shell echo $(PWD) | cut -c1)
endif
endif

target: ~/bin/neo.$(EXT)

~/bin/neo.$(EXT): neo
	chmod 701 neo
	mv neo ~/bin/neo.$(EXT)

clean::
	rm -f ~/bin/neo.$(EXT)
