SRCROOT=.
SRCS=$(shell find $(SRCROOT) -name "*.go")

main: $(SRCS)
	go build -o main main.go

test: main
	/bin/bash -e ./test.sh

clean:
	rm -f main *.o *~ tmp*

.PHONY: test clean