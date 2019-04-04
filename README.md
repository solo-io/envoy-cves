# How to use this tool?

If envoy is in your path, Simply run it!
```
$ envoy-cves
✔ Success! your envoy was tested and is immune to CVE-2019-9901. Make sure that the option normalize_path is turned on in your HCM settings.
✔ Success! your envoy was tested and is immune to CVE-2019-9900
```

If not, provide the path to envoy in a flag:
```
envoy-cves --envoy=/path/to/envoy
✘ Fail! your envoy did not normalize the path - it is vulnerable to CVE-2019-9901
✘ Fail! your envoy accepts nil in headers - it is vulnerable to CVE-2019-9900
```

# What does this tool do?
This tool checks if envoy is vulnerable to CVE-2019-9900 and CVE-2019-9901 by sending envoy
crafted inputs and analyizes how envoy responds to those input. Summarizing the results in a
simple success\fail out.

# How to run this inside a container?
The envoy-cves tool is a single static binary and can easily be used in a docker container.
For example, to test the tool with the official envoy docker image, run this:
```
wget https://github.com/solo-io/envoy-cves/releases/download/v0.1.0/envoy-cves-linux
chmod +x envoy-cves
docker  run -v $PWD/envoy-cves:/bin/envoy-cves --entrypoint=/bin/envoy-cves envoyproxy/envoy
```