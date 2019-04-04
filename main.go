package main

import (
	"fmt"

	"flag"

	"github.com/solo-io/envoy-cves/pkg"
)

var (
	envoy = flag.String("envoy", "envoy", "Path to the envoy binary. defaults to the envoy in your PATH")
	debug = flag.Bool("debug", false, "Output debug info")
)

func main() {
	flag.Parse()

	var er pkg.EnvoyRunner
	er.Debug = *debug
	er.Envoy = *envoy
	er.Log = func(s string) {
		if *debug {
			fmt.Println(s)
		}
	}
	err := pkg.RunChecks(&er)
	if err != nil {
		fmt.Println("Error running CVE checks:", err.Error())
		fmt.Println("This program uses localhost bound ports to test envoy behavior for CVEs.")
		fmt.Println("Try again in case of a temporary error. If error persistent, run with --debug flag and contact the solo team (Join our slack at https://slack.solo.io)")
	}
}
