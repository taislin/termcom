package main

import (
	"fmt"
	"os"

	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/geo"
	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/battle"
)

func main() {
	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	g.RegisterScreen(engine.StateGeoscape, geo.NewGeoscape(g))
	g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g))
	g.RegisterScreen(engine.StateBattlescape, battle.NewBattlescape(g))

	g.Run()
}
