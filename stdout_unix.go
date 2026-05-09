//go:build linux

package main

import (
	"io"
	"os"
)

func getStdout() io.Writer {
	return os.Stdout
}
