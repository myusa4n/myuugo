CFLAGS=-std=c11 -g -static
SRCS=$(wildcard *.go)

main: $(SRCS)
	go build $(SRCS)

test: main
	./test.sh

clean:
	rm -f main *.o *~ tmp*

.PHONY: test clean