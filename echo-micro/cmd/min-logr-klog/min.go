package main

import (
	//"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
)

// Just a dep on logr interface
func main() {
	rootLog := klogr.New()
	rootLog.V(2).Info("test", "a", 1)
}
