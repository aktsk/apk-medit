# medit
![](https://github.com/aktsk/medit/workflows/test/badge.svg)

Medit is a simple memory search and edit tool on Android app.

# Installation 
Download the binary from [GitHub Releases](https://github.com/aktsk/medit/releases/), please push the binary in `/data/local/tmp/` on Android.

```
$ adb push medit /data/local/tmp/medit
medit: 1 file pushed. 29.0 MB/s (3135769 bytes in 0.103s)
```

## How to Build

After the build is complete, if adb is connected, it push the built binary in `/data/local/tmp/` on Android.

```
$ make
GOOS=linux GOARCH=arm64 GOARM=7 go build -o medit
/bin/sh -c "adb push medit /data/local/tmp/medit"
medit: 1 file pushed. 23.7 MB/s (3131205 bytes in 0.126s)
```

# Usage

```
$ adb shell
$ pm list packages # to check <target-package-name>
$ run-as <target-package-name>
$ cp /data/local/tmp/medit ./medit
$ ./medit
```

# Test

```
$ make test
```
