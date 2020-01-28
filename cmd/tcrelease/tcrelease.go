package main

import (
	"fmt"
	"os"

	"github.com/katbyte/tctest/common"
	"github.com/mbfrahry/tcrelease/cmd/tcrelease/cli"
)

func main() {
	if err := cli.Make().Execute(); err != nil {
		common.Log.Errorf(fmt.Sprintf("tcrelease: %v", err))
		os.Exit(1)
	}

	os.Exit(0)
}