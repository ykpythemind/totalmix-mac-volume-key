package main

import (
	"fmt"
	"sync"
)

const (
	DefaultStep = 0.01 // 100 steps from 0 to 1
)

type VolumeState struct {
	mu           sync.Mutex
	level        float32
	muted        bool
	preMuteLevel float32
	step         float32
}

func NewVolumeState() *VolumeState {
	return &VolumeState{
		level: -1, // uninitialized
		step:  DefaultStep,
	}
}

func (v *VolumeState) Initialized() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.level >= 0
}

func (v *VolumeState) Level() float32 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.level
}

func (v *VolumeState) SetLevel(level float32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.level = clamp(level, 0.0, 1.0)
	if v.muted && v.level > 0 {
		v.muted = false
	}
}

func (v *VolumeState) VolumeUp() float32 {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.level < 0 {
		v.level = 0 // safety: if somehow uninitialized
	}
	if v.muted {
		v.muted = false
		v.level = v.preMuteLevel
	}
	v.level = clamp(v.level+v.step, 0.0, 1.0)
	fmt.Printf("Volume Up: %.0f%%\n", v.level*100)
	return v.level
}

func (v *VolumeState) VolumeDown() float32 {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.level < 0 {
		v.level = 0
	}
	if v.muted {
		v.muted = false
		v.level = v.preMuteLevel
	}
	v.level = clamp(v.level-v.step, 0.0, 1.0)
	fmt.Printf("Volume Down: %.0f%%\n", v.level*100)
	return v.level
}

func (v *VolumeState) ToggleMute() float32 {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.level < 0 {
		v.level = 0
	}
	if v.muted {
		v.muted = false
		v.level = v.preMuteLevel
		fmt.Printf("Unmute: %.0f%%\n", v.level*100)
		return v.level
	}
	v.muted = true
	v.preMuteLevel = v.level
	fmt.Println("Mute")
	return 0.0
}

func clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
