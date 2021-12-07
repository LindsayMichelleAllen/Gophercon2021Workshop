//go:build profile
// +build profile

package main

// Add /debug/pprof endpoint
import (
	_ "net/http/pprof"
)
