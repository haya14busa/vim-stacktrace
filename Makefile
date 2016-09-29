all:
	go get -d
	go build -o ./vim-stacktrace

clean:
	rm ./vim-stacktrace

deps:
	go get -d -v -t ./...
	go get github.com/mattn/goveralls
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	go get honnef.co/go/unused/cmd/unused
	go get github.com/haya14busa/go-vimlparser/cmd/vimlparser

check:
	uname -a
	go env
	which -a vim
	vim --version

lint:
	# Go
	go vet ./...
	golint -set_exit_status ./...
	unused ./...
	! gofmt -s -d -l . 2>&1 | grep '^'
	! goimports -l . 2>&1 | grep '^'
	# Vim
	vimlparser **/*.vim > /dev/null

test:
	go test -v -race ./...
