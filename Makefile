.PHONY: build flash monitor

build:
	tinygo build -target=xiao-rp2040 -size=short -o build/firmware.uf2 ./src

flash:
	tinygo flash -target=xiao-rp2040 -size=short ./src

monitor:
	tinygo flash -target=xiao-rp2040 -size=short -monitor ./src 
