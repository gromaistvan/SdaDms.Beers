//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func setConsoleTitle(title string) error {
	handle, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(handle)

	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")
	if err != nil {
		return err
	}
	value, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	_, _, err = syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(value)), 0, 0)
	return err
}
