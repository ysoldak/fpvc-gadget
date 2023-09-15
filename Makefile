TARGET ?= xiao-rp2040 # xiao-ble

.PHONY: build flash monitor

build:
	@mkdir -p build
	tinygo build -target=$(TARGET) -size=short -o build/firmware.uf2 ./src

flash:
	tinygo flash -target=$(TARGET) -size=short ./src

monitor:
	tinygo flash -target=$(TARGET) -size=short -monitor ./src 
