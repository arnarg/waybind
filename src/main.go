package main

import (
	"fmt"
	"os"
	"os/user"

	evdev "github.com/gvalkov/golang-evdev"
	"github.com/jinzhu/configor"
	"gopkg.in/bendahl/uinput.v1"
)

func main() {
	var pressedKeys [KEY_MAX + 1]bool
	var lastPressedKeys [KEY_MAX + 1]bool

	config := Config{}

	// Load config
	user, err := user.Current()
	if err != nil {
		fmt.Println("Could not get current user")
		os.Exit(1)
	}
	configor.New(&configor.Config{ENVPrefix: "WAYBIND"}).Load(&config, "./config.yml", user.HomeDir+"/.config/waybind/config.yml", "/etc/waybind/config.yml")

	// Create virtual keyboard
	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("Waybind Virtual Keyboard"))
	if err != nil {
		fmt.Printf("Could not create virtual keyboard: %s\n", err)
		os.Exit(1)
	}
	defer keyboard.Close()

	// Open real keyboard
	device, err := evdev.Open(config.Device)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get exclusive access to keyboard
	device.Grab()
	defer device.Release()

	for {
		// Read keyboard events
		events, err := device.Read()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Update currently pressed keys
		processEvents(events, pressedKeys[:])

		// Get overlay for pressed keys according to rebind config
		overlayKeys, shouldExit := getRebindOverlay(pressedKeys[:], config.Rebinds)
		if shouldExit {
			fmt.Println("An exit key combination was pressed, exiting.")
			os.Exit(0)
		}

		// Get what's different from pressed key in last iteration
		stateChanges := getStateChanges(pressedKeys[:], overlayKeys, lastPressedKeys[:])

		// Press keys in virtual keyboard if there were any changes
		for _, change := range stateChanges {
			if change.State {
				keyboard.KeyDown(change.Code)
			} else {
				keyboard.KeyUp(change.Code)
			}
		}
	}
}
