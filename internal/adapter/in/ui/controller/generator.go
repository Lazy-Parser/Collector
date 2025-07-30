package controller

import (
	config "Cleopatra/config/service"
	generator "Cleopatra/internal/generator/usecase"
	"Cleopatra/internal/port"
	"context"
)

type GeneratorRepo interface {
	GenerateMexc(ctx context.Context, cfg *config.Config, progress chan generator.Progress) error
}

type Params struct {
	Generator GeneratorRepo
	Logger    port.Logger
}

type GeneratorController struct {
	generator GeneratorRepo
	logger    port.Logger
}

func NewGeneratorController(params *Params) *GeneratorController {
	return &GeneratorController{
		generator: params.Generator,
		logger:    params.Logger,
	}
}

var _ GeneratorRepo = (*generator.Generator)(nil)

func (gen *GeneratorController) TryGenerate(ctx context.Context, cfg *config.Config) {
	progress := make(chan generator.Progress, 100)

	go func() {
		gen.generator.GenerateMexc(ctx, cfg, progress)
	}()

	go func() {
		for data := range progress {
			gen.logger.Log("Max: %d, Current: %d", data.Max, data.Current)
		}
	}()
}
