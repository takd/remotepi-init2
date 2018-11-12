## This looks complicated, but it's really just getting the path of this Makefile
path := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
cwd  := $(shell pwd)


all: boot/pi-init2

boot/pi-init2:
	GOPATH=$(path) GOOS=linux GOARCH=arm go build -o boot/pi-init2 pi-init2

clean:
	rm -f boot/pi-init2

reqs:
	apt update
	apt install -y git golang
	GOPATH=$(path) GOOS=linux GOARCH=arm go get golang.org/x/sys/unix

## This is an experimental and Mac-only shortcut to installing the files onto a mounted card.
## I'm looking for contributions to also support Linux.
install:
	@cd $(path); \
	card_device="$$(diskutil info -all | awk '/Device Identifier/{disk=$$3} /SD Card Reader/{print(disk)}')"; \
	card_path="$$(mount | awk '$$1 ~ "'$${card_device}'" && $$3 ~ "boot" {print $$3}')"; \
	cp -Hv boot/* $${card_path}/

