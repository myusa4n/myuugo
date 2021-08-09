SRCS=$(wildcard *.go)

main: $(SRCS)
	go build -o main $(SRCS)

test: main
	/bin/bash -e ./test.sh

clean:
	rm -f main *.o *~ tmp*

.PHONY: test clean