package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func envoyConfigTemplate(listenerPort, clusterPort uint16) string {
	return fmt.Sprintf(`
node:
 cluster: test
 id: test

static_resources:
  clusters:
  - name: echocluster
    connect_timeout: 0.25s
    type: STATIC
    load_assignment:
      cluster_name: echocluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: %d
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 127.0.0.1, port_value: %d }
    filter_chains:
    - filters:
      - name: envoy.http_connection_manager
        config:
          normalize_path: true
          request_timeout: 1s
          stat_prefix: ingress_http
          codec_type: AUTO
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                route: { cluster: echocluster }
          http_filters:
          - name: envoy.router

admin:
  access_log_path: /dev/null
`, clusterPort, listenerPort)
}

func handlerfunc(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(r.URL.Path))
}

func GetListenerAndPort() (net.Listener, uint16, error) {

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, 0, err
	}

	addr := listener.Addr().String()
	_, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, 0, err
	}

	port, err := strconv.Atoi(portstr)
	return listener, (uint16)(port), err
}

func RunEcho() (uint16, error) {

	handler := http.HandlerFunc(handlerfunc)
	h := &http.Server{Handler: handler}

	listener, port, err := GetListenerAndPort()
	if err != nil {
		if listener != nil {
			listener.Close()
		}
		return 0, err
	}

	go func() {
		h.Serve(listener)
	}()
	return port, nil
}

type EnvoyRunner struct {
	Envoy       string
	Debug       bool
	Log         func(string)
	ClusterPort uint16

	listenerPort uint16
	cmd          *exec.Cmd
	doneChan     <-chan struct{}
}

func (r *EnvoyRunner) Close() error {
	if r.cmd != nil {
		r.cmd.Process.Kill()
		r.cmd = nil
	}
	return nil
}

func (r *EnvoyRunner) Run() error {

	// get a free listener port
	listener, port, err := GetListenerAndPort()
	if listener != nil {
		listener.Close()
	}
	if err != nil {
		return err
	}
	r.listenerPort = port

	args := []string{"--config-yaml", envoyConfigTemplate(r.listenerPort, r.ClusterPort), "--disable-hot-restart", "--allow-unknown-fields"}
	if r.Debug {
		args = append(args, "--log-level", "debug")
	}
	cmd := exec.Command(r.Envoy, args...)
	if r.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	donechan := make(chan struct{})
	r.doneChan = donechan
	// start envoy and wait for it to initialize
	err = cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		cmd.Wait()
		close(donechan)
		r.Log("envoy terminated")
	}()

	r.cmd = cmd
	return nil
}

func (r *EnvoyRunner) WaitForReadyness() error {
	// try to see if envoy is there
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to start envoy.")
		case <-time.After(time.Second):
			_, err := http.Get("http://localhost:" + fmt.Sprintf("%d", r.listenerPort) + "/ready")
			if err == nil {
				return nil
			}
		}
	}
}

func (r *EnvoyRunner) CheckNormalizedPath() (bool, error) {

	resp, err := http.Get("http://localhost:" + fmt.Sprintf("%d", r.listenerPort) + "/folder/../file")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if string(b) == "/file" {
		return true, nil
	}
	if string(b) == "/folder/../file" {
		return false, nil
	}

	return false, fmt.Errorf("unexpected response")
}

func (r *EnvoyRunner) CheckNilErrors() (bool, error) {
	conn, err := net.Dial("tcp", "localhost:"+fmt.Sprintf("%d", r.listenerPort))
	if err != nil {
		// handle error
		return false, fmt.Errorf("can't connect to envoy - check that port 10003 is available")
	}

	// go http doesnt let us send nil headers, so do this the old fashioned way..
	conn.Write([]byte("GET /nilheader HTTP/1.1\r\nHost: localhost:10003\r\nx-test-header: nilsare\000here\r\n\r\n"))

	// envoy should hang up shortly!
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	n, err := conn.Read(make([]byte, 2048))
	if n != 0 {
		return false, nil
	}
	// check that envoy didn't crash..
	select {
	case <-r.doneChan:
		// envoy crashed processing the headers...
		return false, nil
	case <-time.After(time.Second):
	}

	// envoy's still running, and it closed the connection - it has acted properly to our
	// malicious attmepts!
	if err == io.EOF {
		r.Log("nilcheck - connection terminated")
		return true, nil
	}
	if neterr, ok := err.(net.Error); ok {
		if neterr.Timeout() {
			r.Log("nilcheck - network timeout while reading")
			// error is network timeout - envoy didn't close the connection. which means it didn't detect
			// the nil header value
			return false, nil
		}
	}

	// Any other error - not sure...
	return false, err
}
