// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

func Test() error {
	err := sh.Run("mix", "test")
	if err != nil {
		return fmt.Errorf("failed tests: %s", err)
	}

	return nil
}
