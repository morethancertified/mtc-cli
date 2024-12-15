default:
  @just --list

build:
  go build -o bin/mtc main.go

test:
  go test -v ./...

clean:
  rm -rf bin/

tidy:
  go mod tidy

fmt:
  go fmt ./...

run:
  go run main.go

install:
  go install 'github.com/morethancertified/mtc-cli'
  echo "mtc installed"

uninstall:
  rm -f $(go env GOPATH)/bin/mtc-cli
  echo "mtc uninstalled"
