TARGET ?= xiao-ble
FILE = fpvc-gadget_$(TARGET)_$(VERSION).uf2

VERSION := $(shell git describe --tags --always)
LD_FLAGS := -ldflags="-X 'main.Version=$(VERSION)'" # https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications

.PHONY: clean build flash monitor

clean:
	@rm -rf build

build:
	@mkdir -p build
	tinygo build $(LD_FLAGS) -target=$(TARGET) -size=full -o build/$(FILE) ./src

flash:
	tinygo flash $(LD_FLAGS) -target=$(TARGET) -size=short ./src

flash-release:
	cp ./releases/fpvc-gadget_xiao-ble_0.2.1.uf2 /Volumes/XIAO-SENSE/

monitor:
	tinygo monitor -target=$(TARGET)
