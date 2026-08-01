package seed

import (
	"context"
	"fmt"
	"math"
	"time"

	"hstserver/model"
)

// forexSymbol is one instrument and the folder it sits in.
type forexSymbol struct {
	Name   string
	Folder string
}

// currencyNames builds the description, so no row repeats the same words.
var currencyNames = map[string]string{
	"AUD": "Australian Dollar", "CAD": "Canadian Dollar", "CHF": "Swiss Franc",
	"CNH": "Chinese Yuan Offshore", "CZK": "Czech Koruna", "DKK": "Danish Krone",
	"EUR": "Euro", "GBP": "British Pound", "HKD": "Hong Kong Dollar",
	"HUF": "Hungarian Forint", "ILS": "Israeli Shekel", "JPY": "Japanese Yen",
	"MXN": "Mexican Peso", "NOK": "Norwegian Krone", "NZD": "New Zealand Dollar",
	"PLN": "Polish Zloty", "RUB": "Russian Ruble", "SEK": "Swedish Krona",
	"SGD": "Singapore Dollar", "THB": "Thai Baht", "TRY": "Turkish Lira",
	"USD": "US Dollar", "ZAR": "South African Rand",
}

// smallDigitQuotes are quoted with 3 digits instead of 5, the way the platform does.
var smallDigitQuotes = map[string]bool{"JPY": true, "HUF": true}

// forexSymbols are the 60 instruments a fresh install starts with.
var forexSymbols = []forexSymbol{
	// the seven majors
	{"EURUSD", `Forex\Majors`}, {"GBPUSD", `Forex\Majors`}, {"USDJPY", `Forex\Majors`},
	{"USDCHF", `Forex\Majors`}, {"USDCAD", `Forex\Majors`}, {"AUDUSD", `Forex\Majors`},
	{"NZDUSD", `Forex\Majors`},

	// crosses, the majors traded against each other
	{"EURGBP", `Forex\Crosses`}, {"EURJPY", `Forex\Crosses`}, {"EURCHF", `Forex\Crosses`},
	{"EURCAD", `Forex\Crosses`}, {"EURAUD", `Forex\Crosses`}, {"EURNZD", `Forex\Crosses`},
	{"GBPJPY", `Forex\Crosses`}, {"GBPCHF", `Forex\Crosses`}, {"GBPCAD", `Forex\Crosses`},
	{"GBPAUD", `Forex\Crosses`}, {"GBPNZD", `Forex\Crosses`}, {"AUDJPY", `Forex\Crosses`},
	{"AUDCHF", `Forex\Crosses`}, {"AUDCAD", `Forex\Crosses`}, {"AUDNZD", `Forex\Crosses`},
	{"NZDJPY", `Forex\Crosses`}, {"NZDCHF", `Forex\Crosses`}, {"NZDCAD", `Forex\Crosses`},
	{"CADJPY", `Forex\Crosses`}, {"CADCHF", `Forex\Crosses`}, {"CHFJPY", `Forex\Crosses`},

	// exotics, thinner books and wider spreads
	{"USDSEK", `Forex\Exotics`}, {"USDNOK", `Forex\Exotics`}, {"USDDKK", `Forex\Exotics`},
	{"USDPLN", `Forex\Exotics`}, {"USDHUF", `Forex\Exotics`}, {"USDCZK", `Forex\Exotics`},
	{"USDTRY", `Forex\Exotics`}, {"USDZAR", `Forex\Exotics`}, {"USDMXN", `Forex\Exotics`},
	{"USDSGD", `Forex\Exotics`}, {"USDHKD", `Forex\Exotics`}, {"USDCNH", `Forex\Exotics`},
	{"USDTHB", `Forex\Exotics`}, {"USDILS", `Forex\Exotics`}, {"USDRUB", `Forex\Exotics`},
	{"EURSEK", `Forex\Exotics`}, {"EURNOK", `Forex\Exotics`}, {"EURDKK", `Forex\Exotics`},
	{"EURPLN", `Forex\Exotics`}, {"EURHUF", `Forex\Exotics`}, {"EURCZK", `Forex\Exotics`},
	{"EURTRY", `Forex\Exotics`}, {"EURZAR", `Forex\Exotics`}, {"EURMXN", `Forex\Exotics`},
	{"EURSGD", `Forex\Exotics`}, {"EURHKD", `Forex\Exotics`}, {"EURCNH", `Forex\Exotics`},
	{"GBPSEK", `Forex\Exotics`}, {"GBPNOK", `Forex\Exotics`}, {"GBPPLN", `Forex\Exotics`},
	{"GBPTRY", `Forex\Exotics`}, {"GBPZAR", `Forex\Exotics`},
}

// standard settings every seeded forex symbol shares, in platform units.
const (
	seedContractSize = 100000.0

	// volume is a whole number of 1/10000 lot, so multiply lots by 10000
	seedLotUnits = 10000

	seedVolumeMin  = 0.01 * seedLotUnits // 0.01 lot
	seedVolumeMax  = 100 * seedLotUnits  // 100 lots
	seedVolumeStep = 0.01 * seedLotUnits // 0.01 lot

	// a rate of 1 charges the full margin; 0 would ask for none at all
	seedMarginRate = 1.0

	// forex quotes and trades round the clock on weekdays
	seedSessionOpen  = 0
	seedSessionClose = 1440
)

// SeedSymbols fills an empty catalog with the standard forex instruments.
func (s *Seeder) SeedSymbols(ctx context.Context) error {
	var symbols int
	if err := s.DB.DB.QueryRow(ctx, `SELECT count(*) FROM hst.symbols`).Scan(&symbols); err != nil {
		return err
	}
	if symbols > 0 {
		return nil
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	for _, sym := range forexSymbols {
		base, profit := sym.Name[:3], sym.Name[3:]

		digits := int32(5)
		if smallDigitQuotes[profit] {
			digits = 3
		}
		point := math.Pow10(int(-digits))

		var symbolID int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO hst.symbols
			   (symbol, path, description, sector,
			    currency_base, currency_profit, currency_margin,
			    digits, point, multiply, tick_size, tick_flags,
			    trade_mode, calc_mode, exec_mode, gtc_mode,
			    fill_flags, expir_flags, order_flags,
			    contract_size, volume_min, volume_max, volume_step,
			    swap_mode, swap_year_day,
			    margin_initial_buy, margin_initial_sell,
			    margin_initial_buy_limit, margin_initial_sell_limit,
			    margin_initial_buy_stop, margin_initial_sell_stop,
			    margin_initial_buy_stop_limit, margin_initial_sell_stop_limit,
			    margin_maintenance_buy, margin_maintenance_sell,
			    margin_maintenance_buy_limit, margin_maintenance_sell_limit,
			    margin_maintenance_buy_stop, margin_maintenance_sell_stop,
			    margin_maintenance_buy_stop_limit, margin_maintenance_sell_stop_limit,
			    date_created, date_modified)
			 VALUES ($1,$2,$3,$4,$5,$6,$5,$7,$8,$9,$8,$10,
			         $11,$12,$13,$14,$15,$16,$17,
			         $18,$19,$20,$21,$22,$23,
			         $24,$24,$24,$24,$24,$24,$24,$24,
			         $24,$24,$24,$24,$24,$24,$24,$24,
			         $25,$25)
			 RETURNING symbol_id`,
			sym.Name,
			sym.Folder+`\`+sym.Name,
			fmt.Sprintf("%s vs %s", currencyNames[base], currencyNames[profit]),
			model.SymbolSector_currency,
			base, profit,
			digits, point, math.Pow10(int(digits)),
			model.TickFlags_realtime,
			model.TradeMode_full, model.CalcMode_forex, model.ExecMode_market, model.GTCMode_gtc,
			model.FillingFlags_fok|model.FillingFlags_ioc,
			model.ExpirationFlags_gtc|model.ExpirationFlags_day|
				model.ExpirationFlags_specified|model.ExpirationFlags_specified_day,
			model.OrderFlags_market|model.OrderFlags_limit|model.OrderFlags_stop|
				model.OrderFlags_stop_limit|model.OrderFlags_sl|model.OrderFlags_tp|
				model.OrderFlags_closeby,
			seedContractSize, seedVolumeMin, seedVolumeMax, seedVolumeStep,
			model.SwapMode_by_points, model.SwapDays_wednesday,
			seedMarginRate,
			now).
			Scan(&symbolID); err != nil {
			return err
		}

		// without a quote session no price arrives, and without a trade session nothing trades
		for day := int16(1); day <= 5; day++ {
			for _, kind := range []model.SymbolSessionType{
				model.SymbolSessionType_quote, model.SymbolSessionType_trade,
			} {
				if _, err := tx.Exec(ctx,
					`INSERT INTO hst.symbols_sessions (symbol_id, type, day, open, close)
					 VALUES ($1,$2,$3,$4,$5)`,
					symbolID, kind, day, seedSessionOpen, seedSessionClose); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit(ctx)
}
