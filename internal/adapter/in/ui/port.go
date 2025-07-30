package ui_main

import (
	config "Cleopatra/config/service"
	generator "Cleopatra/internal/generator/usecase"
	"context"
)

type TokenServiceRepo interface {
}

type PairServiceRepo interface {
}

type GeneratorRepo interface {
	GenerateMexc(ctx context.Context, cfg *config.Config, progress chan generator.Progress) error
}
