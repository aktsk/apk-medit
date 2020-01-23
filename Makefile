GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=medit
DEVICES:=$(shell adb devices | grep -c 'device$$')

all: build deploy

build:
	GOOS=linux GOARCH=arm64 GOARM=7 $(GOBUILD) -o $(BINARY_NAME)

clean:
	rm $(BINARY_NAME)

deploy:
ifeq ($(DEVICES), 1)
	$(SHELL) -c "adb push $(BINARY_NAME) /data/local/tmp/$(BINARY_NAME)"
else
	@echo 'Android device is not connected....'
endif
