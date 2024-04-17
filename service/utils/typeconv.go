package utils

import "unsafe"

func Bytes2String(bye []byte) string {
	return *(*string)(unsafe.Pointer(&bye))
}

func String2Bytes(strings string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&strings))
	return *(*[]byte)(unsafe.Pointer(&[3]uintptr{x[0], x[1], x[1]}))
}
