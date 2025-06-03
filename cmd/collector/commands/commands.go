package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Lazy-Parser/Collector/internal/core"
	"github.com/Lazy-Parser/Collector/internal/dashboard"
	db "github.com/Lazy-Parser/Collector/internal/database"
	"github.com/Lazy-Parser/Collector/internal/generator"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/cex/mexc"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/evm"
	"github.com/Lazy-Parser/Collector/internal/impl/collector/dex/evm/module"
	manager_cex "github.com/Lazy-Parser/Collector/internal/impl/manager/cex"
	manager_dex "github.com/Lazy-Parser/Collector/internal/impl/manager/dex"
	"github.com/Lazy-Parser/Collector/internal/impl/publisher"
	"github.com/Lazy-Parser/Collector/internal/ui"
	"github.com/Lazy-Parser/Collector/internal/utils"
	"github.com/urfave/cli/v2"
)

func Main(*cli.Context) error {
	ctx := context.TODO()
	ui.CreateUI()
	go ui.GetUI().Run()

	publisher.Init()

	// just show in ui
	go startCex(ctx)
	go startDex(ctx)

	<-ctx.Done()
	publisher.GetPublisher().Close()

	return nil
}

func startCex(ctx context.Context) {
	pairsV3, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Pool: "pancakeswap", Label: "v3"})
	pairsV2, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Pool: "pancakeswap", Label: "v2"})
	pairs := append(pairsV2, pairsV3...)
	ui.GetUI().LogsView(fmt.Sprintf("Length: %d", len(pairs)), "log")

	dataFlow := make(chan core.MexcResponse, 1000)
	tableChan := make(chan core.MexcResponse, 1000)
	publisherChan := make(chan publisher.CexTick, 1000)

	collector := mexc.Mexc{Pool: mexc.CreatePool()}
	manager := manager_cex.CreateManager()
	manager.NewCollector(&collector, pairs)

	go manager.Run(ctx, dataFlow)

	go ui.GetUI().RenderTableCex(tableChan)
	go func() {
		if err := publisher.GetPublisher().PublishStreamCex(publisherChan); err != nil {
			ui.GetUI().LogsView("Failed to push cex tick to nats: "+err.Error(), "error")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(dataFlow)
			close(tableChan)
			close(publisherChan)
			return

		case msg := <-dataFlow:
			if len(msg.Data.Asks) == 0 || len(msg.Data.Bids) == 0 {
				continue
			}

			// show in table
			tableChan <- msg

			// fetch appropriate pair (to get base/quote addresses) and send to publisher
			baseName := strings.Split(msg.Symbols, "_")[0]
			pair, ok := utils.FindPairByBaseName(&pairs, baseName)
			if !ok { // if not found
				slog.Warn("Failed to find pair in CEX main loop by '%s' base name (full '%s')", baseName, msg.Symbols)
			}

			payload := publisher.CexTick{
				Symbols:      msg.Symbols,
				BaseAddress:  pair.BaseToken.Address,
				QuoteAddress: pair.QuoteToken.Address,
				CexName:      "MEXC", // TODO: i should get CexName from collector responce. So i should make custom type, that every collector will fill and add its own collector name
				Bid:          strconv.FormatFloat(msg.Data.Bids[0][0], 'f', -1, 64),
				Ask:          strconv.FormatFloat(msg.Data.Asks[0][0], 'f', -1, 64),
				Timestamp:    msg.TS,
			}

			publisherChan <- payload
		}
	}

}

func startDex(ctx context.Context) {
	allowedPools := []string{"pancakeswap", "uniswap", "sushiswap"}
	// ammEth, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "ethereum", Pool: allowedPools, Label: "v2"})
	ammBsc, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "bsc", Pool: allowedPools, Label: "v2"})
	// clmmEth, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "ethereum", Pool: allowedPools, Label: "v3"})
	clmmBsc, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Network: "bsc", Pool: allowedPools, Label: "v3"})
	quoteChangerPairs, _ := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Type: "quote"})
	allPairs := append(ammBsc, clmmBsc...)
	allPairs = append(allPairs, quoteChangerPairs...)

	dashboardChan := make(chan core.CollectorDexResponse, 100)
	dataFlow := make(chan core.CollectorDexResponse, 100)
	publisherChan := make(chan publisher.DexTick, 100)

	msg := fmt.Sprintf("Lengths of arrays:\n"+
		// "AMM  ETH (v2): %d\n"+
		"AMM  BSC (v2): %d\n"+
		// "CLMM ETH (v3): %d\n"+
		"CLMM BSC (v3): %d\n"+
		"QUOTE CAHNGER: %d\n",
		// len(ammEth),
		len(ammBsc),
		// len(clmmEth),
		len(clmmBsc),
		len(quoteChangerPairs))
	ui.GetUI().LogsView(msg, "log")

	moduleAmm := module.CreateAMM()
	moduleClmm := module.CreateCLMM()

	// moduleAmm.Push(ammEth, "ethereum")
	moduleAmm.Push(ammBsc, "bsc")
	// moduleClmm.Push(clmmEth, "ethereum")
	moduleClmm.Push(clmmBsc, "bsc")

	evmCollector := evm.EVM{}
	evmCollector.Push([]module.EVMModuleImplementation{moduleAmm, moduleClmm})
	manager := manager_dex.New()
	manager.Push(&evmCollector)
	manager.Init(quoteChangerPairs)

	go ui.GetUI().ShowCollectorPrices(dashboardChan)

	go func() {
		if err := manager.Run(ctx, dataFlow); err != nil {
			ui.GetUI().LogsView(err.Error(), "error")
		}
	}()

	go func() {
		if err := publisher.GetPublisher().PublishStreamDex(publisherChan); err != nil {
			ui.GetUI().LogsView("Failed to push dex tick to nats: "+err.Error(), "error")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(dataFlow)
			close(dashboardChan)
			close(publisherChan)
			return

		case msg := <-dataFlow:
			// show in table
			dashboardChan <- msg

			// fetch the appropriate pair (to get base/quote addresses) and send to publisher
			pair, ok := utils.FindPairByAddress(&allPairs, msg.Address)
			if !ok { // pair nor found
				slog.Warn("Failed to find pair in DEX main loop by '%s' pair address", msg.Address)
			}

			payload := publisher.DexTick{
				Network:     pair.Network,
				Pool:        pair.Pool,
				BaseToken:   pair.BaseToken.Address,
				QuoteToken:  pair.QuoteToken.Address,
				Price:       msg.Price.String(),
				PairAddress: msg.Address,
				Timestamp:   msg.Timestamp,
			}

			publisherChan <- payload
		}
	}
}

func Generate(ctx *cli.Context) error {
	generator.Run()

	// try to log all tokens from db
	fmt.Println("Generated tokens:")
	pairs, err := db.GetDB().PairService.GetAllPairs()
	if err != nil {
		return err
	}
	dashboard.ShowPairs(pairs)

	return nil
}

func Table(ctx *cli.Context) error {
	var flag string
	if ctx.Args().Len() == 0 {
		flag = "spairs"
	}
	if ctx.Bool("spairs") {
		flag = "spairs"
	}
	if ctx.Bool("stokens") {
		flag = "stokens"
	}
	if ctx.Bool("clearAll") {
		flag = "clearAll"
	}

	if flag == "spairs" {
		// fetch pairs
		// 195	USD1/WBNB             2        16  0x4a3218606AF9B4728a9F187E1c1a8c07fBC172a9    bsc       pancakeswap  v3     quote
		// 196  JitoSOL/SOL         118         6  Hp53XEtt4S8SvPCXarsLSdGfZBuUr5mMmZmX2DRNXQKp  solana    orca         wp     quote
		// 197  SOL/USDC              6        13  8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj  solana    raydium      CLMM   quote
		// 198  FRAX/USDC            85         4  0x9A834b70C07C81a9fcD6F22E842BF002fBfFbe4D    ethereum  uniswap      v3     quote
		// 199  OETH/WETH           159         8  0x52299416C469843F4e0d54688099966a6c7d720f    ethereum  uniswap      v3     quote
		// 200  WBTC/WETH

		res, err := db.GetDB().PairService.GetAllPairs()

		if err != nil {
			return err
		}

		dashboard.ShowPairs(res)
	} else if flag == "stokens" {
		// fetch tokens
		allowedPools := []string{"pancakeswap", "uniswap", "sushiswap"}
		pairs, err := db.GetDB().PairService.GetAllPairsByQuery(db.PairQuery{Pool: allowedPools})
		//res, err := db.GetDB().TokenService.GetAllTokens()
		if err != nil {
			return err
		}
		tokens := make(map[db.Token]struct{})
		for _, pair := range pairs {
			tokens[pair.QuoteToken] = struct{}{}
		}

		var res []db.Token
		for t := range tokens {
			res = append(res, t)
		}

		dashboard.ShowTokens(res)
	} else if flag == "clearAll" {
		err := db.GetDB().TokenService.ClearTokens()
		if err != nil {
			return err
		}

		err = db.GetDB().PairService.ClearPairs()
		if err != nil {
			return err
		}

		fmt.Println("Database cleaned!")
	}

	return nil
}
