module fpvc-gadget

go 1.22.4

require (
	tinygo.org/x/drivers v0.27.0
	tinygo.org/x/tinydraw v0.4.0
	tinygo.org/x/tinyfont v0.4.0
)

require github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect

require github.com/ysoldak/fpvc-serial-protocol v1.1.1

// replace github.com/ysoldak/fpvc-serial-protocol => ../fpvc-serial-protocol
