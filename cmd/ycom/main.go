package main

import (
	"fmt"
	"os"

	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/geo"
	"github.com/civ13/ycom/internal/save"
)

func main() {
	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	g.OnNewGame = func() {
		gs := geo.NewGeoscape(g)
		g.RegisterScreen(engine.StateGeoscape, gs)
		g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.Base))
		g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.Base))
		g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.Base))
		g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.Base))
		g.SetState(engine.StateGeoscape)
	}

	g.OnContinue = func() {
		sd, err := save.LoadGame(engine.SaveFile)
		if err != nil {
			return
		}
		gs := geo.NewGeoscapeFromSave(g, sd)
		g.RegisterScreen(engine.StateGeoscape, gs)
		g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.Base))
		g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.Base))
		g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.Base))
		g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.Base))
		g.SetState(engine.StateGeoscape)
	}

	g.OnLoadGame = func() {
		var slots []engine.SlotInfo
		for slot := 1; slot <= 10; slot++ {
			sd, err := save.LoadGame(save.SavePath(slot))
			if err != nil {
				continue
			}
			label := engine.FormatSlotLabel(slot, sd.GameTime.Format("2006 Jan 02"), sd.Funds)
			slots = append(slots, engine.SlotInfo{Slot: slot, Label: label})
		}
		picker := engine.NewSlotPickerScreen(g, engine.SlotPickerLoad, slots, func(slot int) {
			sd, err := save.LoadGame(save.SavePath(slot))
			if err != nil {
				return
			}
			gs := geo.NewGeoscapeFromSave(g, sd)
			g.RegisterScreen(engine.StateGeoscape, gs)
			g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.Base))
			g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.Base))
			g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.Base))
			g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.Base))
			g.SetState(engine.StateGeoscape)
		})
		g.PushScreen(picker)
	}

	g.RegisterScreen(engine.StateHelp, engine.NewHelpScreen(g))
	g.RegisterScreen(engine.StateMenu, engine.NewMenuScreen(g))

	g.Run()
}
