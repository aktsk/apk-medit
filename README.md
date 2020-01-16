# medit
Medit is a simple memory search and edit tool on Android app.

# How to Build

```
$ make SHELL=$SHELL
GOOS=linux GOARCH=arm64 GOARM=7 go build -o medit
adb push medit /data/local/tmp/medit
medit: 1 file pushed. 23.7 MB/s (3131205 bytes in 0.126s)
```

# Usage

```
$ adb shell
$ pm list packages
$ run-as <target-package-name>
$ cp /data/local/tmp/medit ./medit
$ ./medit
```
