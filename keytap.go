package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Carbon
#include "keytap_darwin.h"
*/
import "C"
import (
	"runtime"
)

// KeyEvent represents a media key press
type KeyEvent struct {
	KeyCode int // 0=VolumeUp, 1=VolumeDown, 7=Mute
}

var keyEventChan = make(chan KeyEvent, 16)

//export goMediaKeyCallback
func goMediaKeyCallback(keyCode C.int, keyDown C.int) {
	keyEventChan <- KeyEvent{KeyCode: int(keyCode)}
}

// StartEventTap starts the CGEventTap on a dedicated OS thread.
// This function blocks forever.
func StartEventTap() {
	runtime.LockOSThread()
	C.startEventTap()
}
