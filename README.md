# viamgpsd
viam gpsd movement sensor module

to compile for arm64
====
```
env GOOS=linux GOARCH=arm64 make module
viam module upload --platform "linux/arm64" --version <FILL ME IN> module.tar.gz
```
