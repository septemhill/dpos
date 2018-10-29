package main

import (
	"syscall"
)

func checkError(err error) {
	switch err {
	case syscall.ETIMEDOUT:
		return
	}
}
