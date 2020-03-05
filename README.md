# apk-medit

[![GitHub release](https://img.shields.io/github/v/release/aktsk/apk-medit.svg)](https://github.com/aktsk/apk-medit/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/aktsk/apk-medit/blob/master/LICENSE)
![](https://github.com/aktsk/apk-medit/workflows/test/badge.svg)

Apk-medit is a memory search and patch tool for debuggable apk without root & ndk.
It was created for mobile game security testing.

## Demo

This is a demo that uses apk-medit to clear a game that requires one million taps to clear.

<img src="screenshots/terminal.gif" width=680px> <img src="screenshots/demo-app.gif" width=195px>

## Installation

Download the binary from [GitHub Releases](https://github.com/aktsk/apk-medit/releases/), please push the binary in `/data/local/tmp/` on an android device.

```
$ adb push medit /data/local/tmp/medit
medit: 1 file pushed. 29.0 MB/s (3135769 bytes in 0.103s)
```

### How to Build

You can build with make command. It requires a go compiler.
After the build is complete, if adb is connected, it pushes the built binary in `/data/local/tmp/` on an android device.

```
$ make
GOOS=linux GOARCH=arm64 GOARM=7 go build -o medit
/bin/sh -c "adb push medit /data/local/tmp/medit"
medit: 1 file pushed. 23.7 MB/s (3131205 bytes in 0.126s)
```

## Usage

Use the `run-as` command to read files used by the target app, so apk-medit can only be used with apps that have the debuggable attribute enabled.
To enable the debuggable attribute, open `AndroidManifest.xml`, add the following xml attribute in application xml node:

```
android:debuggable="true"
```

After running the `run-as` command, directory is automatically changed. So copy `medit` from `/data/local/tmp/`.
Running `medit` launches an interactive prompt.

```
$ adb shell
$ pm list packages # to check <target-package-name>
$ run-as <target-package-name>
$ cp /data/local/tmp/medit ./medit
$ ./medit
```

### Commands

Here are the commands available in an interactive prompt.

#### find

Search the specified integer on memory.

```
> find 999982
Search UTF-8 String...
Target Value: 999982([57 57 57 57 56 50])
Found: 0!
------------------------
Search Word...
parsing 999982: value out of range
------------------------
Search Double Word...
Target Value: 999982([46 66 15 0])
Found: 1!
Address: 0xe7021f70
```

You can also specify datatype such as string, word, dword, qword.

```
> find dword 999996
Search Double Word...
Target Value: 999996([60 66 15 0])
Found: 1!
Address: 0xe7021f70
```

#### filter

Filter previous search results that match the current search results.

```
> filter 993881
Check previous results of searching dword...
Target Value: 993881([89 42 15 0])
Found: 1!
Address: 0xe7021f70
```

#### patch

Write the specified value on the address found by search.

```
> patch 10
Successfully patched!
```

#### ps

Find the target process and if there is only one, specify it as the target. `ps` runs automatically on startup.

```
> ps
Package: jp.aktsk.tap1000000, PID: 4398
Target PID has been set to 4398.
```


#### attach

If target pid set by `ps`, attach to the target process, stop all processes in the app by ptrace.

```
> attach
Target PID: 4398
Attached TID: 4398
Attached TID: 4405
Attached TID: 4407
Attached TID: 4408
Attached TID: 4410
Attached TID: 4411
Attached TID: 4412
Attached TID: 4413
Attached TID: 4414
Attached TID: 4415
Attached TID: 4418
Attached TID: 4420
Attached TID: 4424
Attached TID: 4429
Attached TID: 4430
Attached TID: 4436
Attached TID: 4437
Attached TID: 4438
Attached TID: 4439
Attached TID: 4440
Attached TID: 4441
Attached TID: 4442
```

If target pid is not set, it can be specified on the command line.

```
> attach <pid>
```

#### detach

Detach from the attached process.

```
> detach
Detached TID: 4398
Detached TID: 4405
Detached TID: 4407
Detached TID: 4408
Detached TID: 4410
Detached TID: 4411
Detached TID: 4412
Detached TID: 4413
Detached TID: 4414
Detached TID: 4415
Detached TID: 4418
Detached TID: 4420
Detached TID: 4424
Detached TID: 4429
Detached TID: 4430
Detached TID: 4436
Detached TID: 4437
Detached TID: 4438
Detached TID: 4439
Detached TID: 4440
Detached TID: 4441
Detached TID: 4442
```

#### dump

Display memory dump like hexdump.

```
> dump 0xf0aee000 0xf0aee300
Address range: 0xf0aee000 - 0xf0aee300
----------------------------------------------
00000000  34 32 20 61 6e 73 77 65  72 20 28 74 6f 20 6c 69  |42 answer (to li|
00000010  66 65 20 74 68 65 20 75  6e 69 76 65 72 73 65 20  |fe the universe |
00000020  65 74 63 7c 33 29 0a 33  31 34 20 70 69 0a 31 30  |etc|3).314 pi.10|
00000030  30 33 20 61 75 64 69 74  64 20 28 61 76 63 7c 33  |03 auditd (avc|3|
00000040  29 0a 31 30 30 34 20 63  68 61 74 74 79 20 28 64  |).1004 chatty (d|
00000050  72 6f 70 70 65 64 7c 33  29 0a 31 30 30 35 20 74  |ropped|3).1005 t|
00000060  61 67 5f 64 65 66 20 28  74 61 67 7c 31 29 2c 28  |ag_def (tag|1),(|
00000070  6e 61 6d 65 7c 33 29 2c  28 66 6f 72 6d 61 74 7c  |name|3),(format||
00000080  33 29 0a 31 30 30 36 20  6c 69 62 6c 6f 67 20 28  |3).1006 liblog (|
00000090  64 72 6f 70 70 65 64 7c  31 29 0a 32 37 31 38 20  |dropped|1).2718 |
000000a0  65 0a 32 37 31 39 20 63  6f 6e 66 69 67 75 72 61  |e.2719 configura|
000000b0  74 69 6f 6e 5f 63 68 61  6e 67 65 64 20 28 63 6f  |tion_changed (co|
000000c0  6e 66 69 67 20 6d 61 73  6b 7c 31 7c 35 29 0a 32  |nfig mask|1|5).2|
000000d0  37 32 30 20 73 79 6e 63  20 28 69 64 7c 33 29 2c  |720 sync (id|3),|
000000e0  28 65 76 65 6e 74 7c 31  7c 35 29 2c 28 73 6f 75  |(event|1|5),(sou|
000000f0  72 63 65 7c 31 7c 35 29  2c 28 61 63 63 6f 75 6e  |rce|1|5),(accoun|
```

#### exit

To exit medit, use the `exit` command or `Ctrl-D`.

```
> exit
Bye!
```

## Test

You can run test codes with make command.

```
$ make test
```

## License

MIT License
