//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/datafs"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/geo"
	"github.com/taislin/termcom/internal/mapgen"
	"github.com/taislin/termcom/internal/save"
)

func main() {
	c := make(chan struct{})

	js.Global().Set("termcomInit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			datafs.Set(embeddedFS())
			engine.LoadConfig()
			initMapTiles()
			g := engine.NewGameWASM()
			setupGame(g)
			g.Run()
		}()
		return nil
	}))

	<-c
}

func initMapTiles() {
	if err := mapgen.Init(); err != nil {
		// WASM: data files may not be available, skip
	}
	battle.InitCustomTiles()
	data.NewAlienSpriteRegistry().RebuildFromTemplates(
		mapgen.ToTemplateData("head"),
		mapgen.ToTemplateData("eye"),
		mapgen.ToTemplateData("torso"),
		mapgen.ToTemplateData("leg"),
		mapgen.ToTemplateData("weapon"),
	)
}

func setupGame(g *engine.Game) {
	g.RegisterScreen(engine.StateMenu, engine.NewMenuScreen(g))
	g.RegisterScreen(engine.StateLanguageSelect, engine.NewLanguageSelectScreen(g))
	g.RegisterScreen(engine.StateHelp, engine.NewHelpScreen(g, engine.StateGeoscape))

	g.OnNewGame = func() {
		seedScreen := engine.NewSeedScreen(g, func(seed int64) {
			_ = seed
			picker := engine.NewDifficultyScreen(g, func(difficulty int) {
				gs := geo.NewGeoscape(g)
				g.RegisterScreen(engine.StateGeoscape, gs)
				g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
				g.SetState(engine.StateGeoscape)
			})
			g.PushScreen(picker)
		})
		g.PushScreen(seedScreen)
	}

	g.OnContinue = func() {
		sd, err := save.LoadGame(engine.SaveFile)
		if err != nil {
			return
		}
		gs := geo.NewGeoscapeFromSave(g, sd)
		g.RegisterScreen(engine.StateGeoscape, gs)
		g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
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
			g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
			g.SetState(engine.StateGeoscape)
		})
		g.PushScreen(picker)
	}

	g.OnCustomBattle = func() {
		screen := engine.NewCustomBattleScreen(g, func(entry engine.CustomBattleEntry) {
			launchCustomBattle(g, entry.FilePath)
		})
		g.PushScreen(screen)
	}
}

func launchCustomBattle(g *engine.Game, path string) {
	// WASM: custom battles not supported (file I/O unavailable)
}
