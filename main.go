package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sendPort := flag.Int("send-port", 7001, "TotalMix FX OSC incoming port")
	recvPort := flag.Int("recv-port", 9001, "TotalMix FX OSC outgoing port (listen)")
	step := flag.Float64("step", 0.02, "Volume step per key press (0.0-1.0)")
	flag.Parse()

	vol := NewVolumeState()
	vol.step = float32(*step)

	// OSC client for sending to TotalMix FX
	client, err := NewOSCClient("127.0.0.1", *sendPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create OSC client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// OSC listener for state sync from TotalMix FX
	listener, err := NewOSCListener(*recvPort, vol)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not start OSC listener on port %d: %v\n", *recvPort, err)
		fmt.Fprintln(os.Stderr, "Volume sync from TotalMix FX will not work, but sending commands will still work.")
	} else {
		defer listener.Close()
		go listener.Listen()
	}

	fmt.Printf("mac_audio_vol: Sending OSC to 127.0.0.1:%d, listening on :%d\n", *sendPort, *recvPort)
	fmt.Println("Waiting for media key events... (Ctrl+C to quit)")
	fmt.Println("Make sure Accessibility permission is granted and TotalMix FX OSC is enabled.")

	// Handle key events in a goroutine
	go func() {
		for ev := range keyEventChan {
			var level float32
			switch ev.KeyCode {
			case 0: // NX_KEYTYPE_SOUND_UP
				level = vol.VolumeUp()
			case 1: // NX_KEYTYPE_SOUND_DOWN
				level = vol.VolumeDown()
			case 7: // NX_KEYTYPE_MUTE
				level = vol.ToggleMute()
			default:
				continue
			}
			if err := client.SendFloat("/1/mastervolume", level); err != nil {
				fmt.Fprintf(os.Stderr, "OSC send error: %v\n", err)
			}
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start event tap on main goroutine (blocks on CFRunLoop)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	StartEventTap()
}
