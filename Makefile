GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=medit

all: build deploy

build:
	GOOS=linux GOARCH=arm64 GOARM=7 $(GOBUILD) -o $(BINARY_NAME)

clean:
	rm $(BINARY_NAME)

deploy:
	$(SHELL) -c "adb push $(BINARY_NAME) /data/local/tmp/$(BINARY_NAME)"
