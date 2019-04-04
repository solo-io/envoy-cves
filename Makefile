envoycve: main.go pkg/envoy.go pkg/runchecks.go
	CGO_ENABLED=0 go build -o $@ main.go
	strip $@