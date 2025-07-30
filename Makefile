.PHONY: build clean all

BINARY_NAME=go-hurobot
BINARY_PATH=bin

all: build

$(BINARY_PATH):
	mkdir -p $(BINARY_PATH)

build: $(BINARY_PATH)
	CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BINARY_PATH)/$(BINARY_NAME)
	# go build -o $(BINARY_PATH)/$(BINARY_NAME) .

clean:
	rm -rf $(BINARY_PATH)
