package ui_main

import (
	userinterface "Cleopatra/internal/adapter/in/ui/view"
	"Cleopatra/internal/port"
	"fmt"
	"os/exec"
)

type Params struct {
	Logger       port.Logger
	TokenService TokenServiceRepo
	PairService  PairServiceRepo
	Generator    GeneratorRepo
}

func Run(params *Params) error {
	// create ui
	ui := userinterface.Create(params.Logger, params.Generator)

	// main loop. It blocks main
	if err := ui.Run(); err != nil {
		ui.Stop()
		exec.Command("clear") // for unix
		return fmt.Errorf("ui: %v", err)
	}

	return nil
}
