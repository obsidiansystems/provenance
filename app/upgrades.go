package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

var (
	noopHandler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Applying no-op upgrade plan for release " + plan.Name)
		return versionMap, nil
	}
)

type appUpgradeHandler = func(*App, sdk.Context, upgradetypes.Plan) (module.VersionMap, error)

type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler appUpgradeHandler
}

var handlers = map[string]appUpgrade{
	"v0.2.0": {},
	"v0.2.1": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.MarkerKeeper.SetParams(ctx, markertypes.DefaultParams())
			return app.UpgradeKeeper.GetModuleVersionMap(ctx), nil
		},
	},
	"v0.3.0": {},
	"v1.0.0": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.NameKeeper.ConvertLegacyAmino(ctx)
			app.AttributeKeeper.ConvertLegacyAmino(ctx)
			return app.UpgradeKeeper.GetModuleVersionMap(ctx), nil
		},
	},
	"v1.1.1":   {},
	"amaranth": {}, // associated with v1.2.x upgrades in testnet, mainnet
	"bluetiful": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			// Force default denom metadata for the bond denom
			app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
				Description: "Hash is the staking token of the Provenance Blockchain",
				Base:        "nhash",
				Display:     "hash",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "nhash",
						Exponent: 0,
						Aliases:  []string{},
					},
					{
						Denom:    "hash",
						Exponent: 9,
						Aliases:  []string{},
					},
				},
			})
			// Force default unrestricted denom for markers to limit min length of 8 and allow ['.','-'] as separators.
			app.MarkerKeeper.SetParams(ctx, markertypes.Params{
				UnrestrictedDenomRegex: `[a-zA-Z][a-zA-Z0-9\-\.]{7,64}`,
			})
			return app.UpgradeKeeper.GetModuleVersionMap(ctx), nil
		},
	},
	"citrine": {},
	"desert":  {},
	"eigengrau": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

			nhashName := "Hash"
			nhashSymbol := "HASH"
			nhash, found := app.BankKeeper.GetDenomMetaData(ctx, "nhash")
			if found {
				nhash.Name = nhashName
				nhash.Symbol = nhashSymbol
			} else {
				nhash = banktypes.Metadata{
					Description: "Hash is the staking token of the Provenance Blockchain",
					Base:        "nhash",
					Display:     "hash",
					Name:        nhashName,
					Symbol:      nhashSymbol,
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "nhash",
							Exponent: 0,
							Aliases:  []string{},
						},
						{
							Denom:    "hash",
							Exponent: 9,
							Aliases:  []string{},
						},
					},
				}
			}
			app.BankKeeper.SetDenomMetaData(ctx, nhash)

			fromVM := map[string]uint64{
				"ibc":       1,
				"attribute": 1,
				"marker":    1,
				"metadata":  1,
				"name":      1,
			}
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	},
	// TODO - Add new upgrade definitions here.
}

func InstallCustomUpgradeHandlers(app *App) {
	// Register all explicit appUpgrades
	for name, upgrade := range handlers {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = noopHandler
		} else {
			ref := upgrade
			handler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				return ref.Handler(app, ctx, plan)
			}
		}
		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}
}

// CustomUpgradeStoreLoader provides upgrade handlers for store and application module upgrades at specified versions
func CustomUpgradeStoreLoader(app *App, info storetypes.UpgradeInfo) baseapp.StoreLoader {
	// Current upgrade info is empty or we are at the wrong height, skip this.
	if info.Name == "" || info.Height-1 != app.LastBlockHeight() {
		return nil
	}
	// Find the upgrade handler that matches this currently executing upgrade.
	for name, upgrade := range handlers {
		// If the plan is executing this block, set the store locator to create any
		// missing modules, delete unused modules, or rename any keys required in the plan.
		if info.Name == name && !app.UpgradeKeeper.IsSkipHeight(info.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added:   upgrade.Added,
				Renamed: upgrade.Renamed,
				Deleted: upgrade.Deleted,
			}

			if isEmptyUpgrade(storeUpgrades) {
				app.Logger().Info("No store upgrades required",
					"plan", name,
					"height", info.Height,
				)
				return nil
			}

			app.Logger().Info("Store upgrades",
				"plan", name,
				"height", info.Height,
				"upgrade.added", upgrade.Added,
				"upgrade.deleted", upgrade.Deleted,
				"upgrade.renamed", upgrade.Renamed,
			)
			return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
		}
	}
	return nil
}

func isEmptyUpgrade(upgrades storetypes.StoreUpgrades) bool {
	return len(upgrades.Renamed) == 0 && len(upgrades.Deleted) == 0 && len(upgrades.Added) == 0
}
