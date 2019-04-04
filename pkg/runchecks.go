package pkg

import (
	"fmt"
)

func RunChecks(er *EnvoyRunner) error {
	port, err := RunEcho()
	if err != nil {
		er.Log("Error running echo server. Is port 10004 available?")
		return err
	}

	er.ClusterPort = port
	err = er.Run()
	if err != nil {
		return err
	}
	defer er.Close()

	err = er.WaitForReadyness()
	if err != nil {
		er.Log("Envoy is not ready to accept connections")
		return err
	}

	b, err := er.CheckNormalizedPath()
	if err != nil {
		er.Log("Error checking path normalization")
		return err
	}
	if b {
		fmt.Println("✔ Success! your envoy was tested and it is immune to CVE-2019-9901. Make sure the option normalize_path is turned on in your HCM settings.")
	} else {
		fmt.Println("✘ Fail! your envoy did not normalize the path - it is vulnerable to CVE-2019-9901. If you are using a recent version of envoy, make sure the option normalize_path is turned on in your HCM settings.")
	}

	// nil checks may crash envoy so do them last.
	headers, err := er.CheckNilErrors()
	if err != nil {
		er.Log("Error checking NIL headers")
		return err
	}

	if headers {
		fmt.Println("✔ Success! your envoy was tested and it is immune to CVE-2019-9900")
	} else {
		fmt.Println("✘ Fail! your envoy accepts NIL in headers - it is vulnerable to CVE-2019-9900")
	}

	return nil
}
