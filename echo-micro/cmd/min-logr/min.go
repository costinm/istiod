package main

import (
	"github.com/go-logr/logr"
)

// Just a dep on logr interface
func main() {
	rootLog := logr.Discard()
	rootLog.V(4).Info("test", "a", 1)
}
