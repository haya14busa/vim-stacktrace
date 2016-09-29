all:
	go get -d
	go build -o ./vim-stacktrace

clean:
	rm ./vim-stacktrace
