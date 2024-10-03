NAME := stc
GIT_TAG=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.GitTag=$(GIT_TAG)"

all: $(NAME)
$(NAME): *.go
	go build $(LDFLAGS) -o $(NAME) .

cross:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-linux .
	GOOS=linux GOARCH=arm go build $(LDFLAGS) -a -o $(NAME)-arm-linux .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -a -o $(NAME)-arm64-linux .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-macos .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -a -o $(NAME)-arm64-macos .
	GOOS=freebsd GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-freebsd .
	GOOS=openbsd GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-openbsd .
	GOOS=netbsd GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-netbsd .
	GOOS=solaris GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-adm64-solaris .
	GOOS=aix GOARCH=ppc64 go build $(LDFLAGS) -a -o $(NAME)-ppc64-aix .
	GOOS=plan9 GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-plan9 .
	GOOS=plan9 GOARCH=arm go build $(LDFLAGS) -a -o $(NAME)-arm-plan9 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -a -o $(NAME)-amd64-win64.exe
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -a -o $(NAME)-arm64-win64.exe
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -a -o $(NAME)-arm64-win32.exe

clean:
	rm -f stc stc-*
