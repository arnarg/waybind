package main

import (
	"bytes"

	evdev "github.com/gvalkov/golang-evdev"
)

const (
	PRESSED     = 1
	NOT_PRESSED = 0
)

// Goes through a list of rebind configs and compares it to the current state of
// pressed keys. It creates a slice with keys that should be pressed or not pressed
// and can be overlaid over top of the currently pressed keys. Unaffected keys are
// represented by 255, pressed by 1 and forcably unpressed 0.
func getRebindOverlay(pressedKeys []bool, rebinds []Rebind) (overlayKeys []byte, shouldExit bool) {
	overlayKeys = bytes.Repeat([]byte{255}, KEY_MAX+1)

	for _, rebind := range rebinds {
		shouldExit = processRebind(&rebind, pressedKeys, overlayKeys)
		if shouldExit {
			break
		}
	}
	return
}

// Goes through a single rebind config and decides which part gets chosen.
func processRebind(rebind *Rebind, pressedKeys []bool, overlayKeys []byte) bool {
	key, ok := ecodes[rebind.From]
	if ok && pressedKeys[key] && overlayKeys[key] > 1 {
		// Check if any modifiers are pressed
		found, shouldExit := processModifiers(rebind, pressedKeys, overlayKeys)
		if shouldExit {
			return true
		}
		if found {
			return false
		}

		// Unbind
		if rebind.Unbind {
			overlayKeys[key] = NOT_PRESSED
			return false
		}

		// Rebind standard
		toKey, ok := ecodes[rebind.To]
		if ok {
			overlayKeys[key] = NOT_PRESSED
			overlayKeys[toKey] = PRESSED
		}
	}
	return false
}

// Goes through a list of modifiers for a rebind and picks the first one it finds.
func processModifiers(rebind *Rebind, pressedKeys []bool, overlayKeys []byte) (found bool, shouldExit bool) {
	key, ok := ecodes[rebind.From]
	if !ok {
		return
	}
	for _, mod := range rebind.Modifiers {
		modKey, ok := ecodes[mod.Modifier]
		if ok && pressedKeys[modKey] {
			// If skip, do nothing
			if mod.To == "SKIP" {
				found = true
				return
			}
			// If Exit, exit
			if mod.To == "EXIT" {
				shouldExit = true
				return
			}
			// Rebind with modifier
			toKey, ok := ecodes[mod.To]
			if ok {
				overlayKeys[key] = NOT_PRESSED
				overlayKeys[modKey] = NOT_PRESSED
				overlayKeys[toKey] = PRESSED
				found = true
				return
			}
		}
	}
	return
}

// Goes through a list of evdev events and updates the currently pressed keys.
func processEvents(e []evdev.InputEvent, pressedKeys []bool) {
	for i := range e {
		event := e[i]
		if event.Type == evdev.EV_KEY {
			pressedKeys[event.Code] = event.Value > 0
		}
	}
}

type StateChange struct {
	Code  int
	State bool
}

// Overlays rebinds on top of pressed keys and compares to last iteration's pressed
// keys to get a list of state changes.
func getStateChanges(pressedKeys []bool, overlayKeys []byte, lastPressedKeys []bool) (changes []StateChange) {
	for i := range pressedKeys {
		var keyState bool
		if overlayKeys[i] < 2 {
			keyState = overlayKeys[i] == 1
		} else {
			keyState = pressedKeys[i]
		}
		if keyState != lastPressedKeys[i] {
			changes = append(changes, StateChange{Code: i, State: keyState})
			lastPressedKeys[i] = keyState
		}
	}
	return
}
