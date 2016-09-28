all:
	go get -d
	go build -o ./vim-callstack

clean:
	rm ./vim-callstack
