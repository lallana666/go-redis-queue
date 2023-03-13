package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/golang/glog"

	"denggotech.cn/heque/heque/cmd/heque-apiserver/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := app.NewAPIServerCommand()

	// add go flags into command's flags, since the glog library uses go flag
	command.Flags().AddGoFlagSet(flag.CommandLine)
	defer glog.Flush()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
