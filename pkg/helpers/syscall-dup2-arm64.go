//go:build arm64
// +build arm64

package helpers

import "syscall"

func PortableDup2(oldfd int, newfd int) error {
	return syscall.Dup3(oldfd, newfd, 0) //nolint:wrapcheck
}
