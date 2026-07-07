package main

import (
	"fmt"
	"os"

	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/battle"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/geo"
)

func main() {
	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	gs := geo.NewGeoscape(g)
	g.RegisterScreen(engine.StateGeoscape, gs)
	g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.Base))
	g.RegisterScreen(engine.StateBattlescape, battle.NewBattlescape(g, nil, ""))
	g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.Base))
	g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.Base))
	g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.Base))

	g.Run()
}
