# How to use this tool?

If envoy is in your path, Simply run it!
```
$ envoycve
✔ Success! your envoy was tested and is immune to CVE-2019-9901. Make sure that normalize_path is turned on in your HCM settings.
✔ Success! your envoy was tested and is immune to CVE-2019-9900
```

If not, provide the path to envoy in a flag:
```
envoycve --envoy=/path/to/envoy
✘ Fail! your envoy did not normalize the path - it is vulnerable to CVE-2019-9901
✘ Fail! your envoy accepts nil in headers - it is vulnerable to CVE-2019-9900
```

# What does this tool do?
This tool checks if envoy is vulnerable to CVE-2019-9900 and CVE-2019-9901 by sending envoy
crafted inputs and analyizes how envoy responds to those input. Summarizing the results in a
simple success\fail out.

# How to run this inside a container?
The envoycve tool is a single static binary and can easily be used in a docker container.
For example, to test the tool with the official envoy docker image, run this:
```
wget FILL IN URL FOR envoycve
chmod +x envoycve
docker  run -v $PWD/envoycve:/bin/envoycve --entrypoint=/bin/envoycve envoyproxy/envoy
```