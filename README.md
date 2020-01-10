# medit
Medit is a simple memory search and edit tool on Android app.

# How to Build

```
$ GOOS=linux GOARCH=arm64 GOARM=7 go build
```

# Usage

```
$ adb push ./medit /data/local/tmp/medit
$ adb shell
$ pm list packages
$ run-as <target-package-name>
$ cp /data/local/tmp/medit ./medit
$ ./medit <pid>
```
