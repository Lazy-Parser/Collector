package api_test

import (
	"path/filepath"

	"github.com/Lazy-Parser/Collector/api"
	"github.com/Lazy-Parser/Collector/config"
)

func ExampleNewMexcApi() {
	path := filepath.Join("some", "path", "to", "the", ".env")
	cfgService, err := config.NewConfig(path)
	if err != nil {
		// do not panic in real app
		panic(err)
	}
	
	mexcApiService := api.NewMexcApi(cfgService)

	// do something with service
	_ = mexcApiService
}

func ExampleNewDexscreenerApi() {
	path := filepath.Join("some", "path", "to", "the", ".env")
	cfgService, err := config.NewConfig(path)
	if err != nil {
		// do not panic in real app
		panic(err)
	}
	
	dsApiService := api.NewDexscreenerApi(cfgService)

	// do something with service
	_ = dsApiService
}

func ExampleNewCoingeckoApi() {
	path := filepath.Join("some", "path", "to", "the", ".env")
	cfgService, err := config.NewConfig(path)
	if err != nil {
		// do not panic in real app
		panic(err)
	}
	
	dsApiService := api.NewCoingeckoApi(cfgService)

	// do something with service
	_ = dsApiService
}