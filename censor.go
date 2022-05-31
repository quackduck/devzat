package main

import "unsafe"

// #cgo LDFLAGS: librustrict_devzat.a -Wl,--no-as-needed -ldl 
// void censor(char* str);
// void free(void* p);
import "C"

func rmBadWords(text string) string {
    cText := C.CString(text)
    defer C.free(unsafe.Pointer(cText))
    C.censor(cText)
    ret := C.GoString(cText)
	return ret
}

