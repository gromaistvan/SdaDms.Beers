//go:build !windows

package main

import (
	"fmt"
)

func setConsoleTitle(title string) error {
	return fmt.Errorf("Not implemented")
}
