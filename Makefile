all: sha

envoy-cves: main.go pkg/envoy.go pkg/runchecks.go
	CGO_ENABLED=0 go build -o $@ main.go
	strip $@

envoy-cves-linux: main.go pkg/envoy.go pkg/runchecks.go
	GOOS=linux CGO_ENABLED=0 go build -o $@ main.go

envoy-cves-darwin: main.go pkg/envoy.go pkg/runchecks.go
	GOOS=darwin CGO_ENABLED=0 go build -o $@ main.go

envoy-cves-linux.sha256: envoy-cves-linux
	sha256sum $< > $@

envoy-cves-darwin.sha256: envoy-cves-darwin
	sha256sum $< > $@

.PHONY: sha
sha: envoy-cves-linux.sha256 envoy-cves-darwin.sha256

clean:
	rm envoy-cves-linux.sha256 envoy-cves-darwin.sha256 envoy-cves envoy-cves-linux envoy-cves-darwin