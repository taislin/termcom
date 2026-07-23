//go:build js && wasm

package engine

import (
	"syscall/js"

	"github.com/gdamore/tcell/v3"
)

var singleton *Game

type wasmBridge struct {
	g       *Game
	fbReady bool
}

var bridge = &wasmBridge{}

func NewGameWASM() *Game {
	LoadConfig()
	cols, rows := 120, 40
	scr := NewScreenRawWASM(cols, rows)

	g := newGameWithScreen(scr, StateMenu)

	g.OnScreenChange = func() {
		bridge.fbReady = false
	}

	singleton = g
	bridge.g = g

	registerCallbacks()

	return g
}

func registerCallbacks() {
	js.Global().Set("termcomInjectKey", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		key := args[0].String()
		ev := parseWasmKey(key)
		if ev != nil && bridge.g != nil {
			bridge.g.InjectKey(ev)
		}
		return nil
	}))

	js.Global().Set("termcomInjectMouse", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return nil
		}
		x := args[0].Int()
		y := args[1].Int()
		action := args[2].String()

		var btn tcell.ButtonMask = tcell.ButtonNone
		switch action {
		case "left":
			btn = tcell.Button1
		case "right":
			btn = tcell.Button2
		case "scroll_up":
			btn = tcell.WheelUp
		case "scroll_down":
			btn = tcell.WheelDown
		}

		ev := tcell.NewEventMouse(x, y, btn, tcell.ModNone)
		if bridge.g != nil {
			bridge.g.InjectMouse(ev)
		}
		return nil
	}))

	js.Global().Set("termcomInjectResize", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			return nil
		}
		cols := args[0].Int()
		rows := args[1].Int()
		if bridge.g != nil {
			bridge.g.InjectResizeWasm(cols, rows)
		}
		return nil
	}))

	js.Global().Set("termcomGetFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if bridge.g == nil || bridge.g.screen == nil {
			return nil
		}
		fb := bridge.g.screen.fb
		data := fb.MarshalBinary()
		w := fb.Width()
		h := fb.Height()
		dirty := !bridge.fbReady
		bridge.fbReady = true
		arr := js.Global().Get("Uint8Array").New(len(data))
		js.CopyBytesToJS(arr, data)
		result := js.Global().Get("Object").New()
		result.Set("data", arr)
		result.Set("w", w)
		result.Set("h", h)
		result.Set("dirty", dirty)
		return result
	}))
}

func parseWasmKey(key string) *tcell.EventKey {
	var k tcell.Key
	r := ""
	switch key {
	case "Enter":
		k = tcell.KeyEnter
	case "Escape":
		k = tcell.KeyEscape
	case "Backspace":
		k = tcell.KeyBackspace
	case "Delete":
		k = tcell.KeyDelete
	case "Tab":
		k = tcell.KeyTab
	case "ArrowUp":
		k = tcell.KeyUp
	case "ArrowDown":
		k = tcell.KeyDown
	case "ArrowLeft":
		k = tcell.KeyLeft
	case "ArrowRight":
		k = tcell.KeyRight
	case "Home":
		k = tcell.KeyHome
	case "End":
		k = tcell.KeyEnd
	case "PageUp":
		k = tcell.KeyPgUp
	case "PageDown":
		k = tcell.KeyPgDn
	case "F1":
		k = tcell.KeyF1
	case "F2":
		k = tcell.KeyF2
	case "F3":
		k = tcell.KeyF3
	case "F4":
		k = tcell.KeyF4
	case "F5":
		k = tcell.KeyF5
	case "F6":
		k = tcell.KeyF6
	case "F7":
		k = tcell.KeyF7
	case "F8":
		k = tcell.KeyF8
	case "F9":
		k = tcell.KeyF9
	case "F10":
		k = tcell.KeyF10
	case "F11":
		k = tcell.KeyF11
	case "F12":
		k = tcell.KeyF12
	default:
		if len(key) == 1 {
			r = key
			k = tcell.KeyRune
		} else {
			return nil
		}
	}
	return tcell.NewEventKey(k, r, tcell.ModNone)
}

func (g *Game) InjectResizeWasm(cols, rows int) {
	if ws, ok := g.screen.screen.(*wasmScreen); ok {
		ws.SetSize(cols, rows)
	}
	g.screen.UpdateSize()
	ev := tcell.NewEventResize(cols, rows)
	select {
	case g.keyChan <- ev:
	default:
	}
}
