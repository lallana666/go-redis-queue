package apiserver

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
