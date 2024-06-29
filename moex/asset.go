package moex

import (
	"fmt"
	"log"
	"log/slog"
)

type AssetInfo []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Securities []struct {
		Secid            string  `json:"SECID"`
		Boardid          string  `json:"BOARDID"`
		Shortname        string  `json:"SHORTNAME"`
		Secname          string  `json:"SECNAME"`
		Prevsettleprice  float64 `json:"PREVSETTLEPRICE"`
		Decimals         int     `json:"DECIMALS"`
		Minstep          float64 `json:"MINSTEP"`
		Lasttradedate    string  `json:"LASTTRADEDATE"`
		Lastdeldate      string  `json:"LASTDELDATE"`
		Sectype          string  `json:"SECTYPE"`
		Latname          string  `json:"LATNAME"`
		Assetcode        string  `json:"ASSETCODE"`
		Prevopenposition int     `json:"PREVOPENPOSITION"`
		Lotvolume        int     `json:"LOTVOLUME"`
		Initialmargin    float64 `json:"INITIALMARGIN"`
		Highlimit        float64 `json:"HIGHLIMIT"`
		Lowlimit         float64 `json:"LOWLIMIT"`
		Stepprice        float64 `json:"STEPPRICE"`
		Lastsettleprice  float64 `json:"LASTSETTLEPRICE"`
		Prevprice        float64 `json:"PREVPRICE"`
		Imtime           string  `json:"IMTIME"`
		Buysellfee       float64 `json:"BUYSELLFEE"`
		Scalperfee       float64 `json:"SCALPERFEE"`
		Negotiatedfee    float64 `json:"NEGOTIATEDFEE"`
		Exercisefee      float64 `json:"EXERCISEFEE"`
	} `json:"securities,omitempty"`
}

type AssetDescription []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Boards []struct {
		Secid        string `json:"secid"`
		Boardid      string `json:"boardid"`
		Title        string `json:"title"`
		BoardGroupID int    `json:"board_group_id"`
		MarketID     int    `json:"market_id"`
		Market       string `json:"market"`
		EngineID     int    `json:"engine_id"`
		Engine       string `json:"engine"`
		IsTraded     int    `json:"is_traded"`
		Decimals     int    `json:"decimals"`
		HistoryFrom  string `json:"history_from"`
		HistoryTill  string `json:"history_till"`
		ListedFrom   string `json:"listed_from"`
		ListedTill   string `json:"listed_till"`
		IsPrimary    int    `json:"is_primary"`
		Currencyid   any    `json:"currencyid"`
	} `json:"boards,omitempty"`
}

// ///////////////////////////////////////////////////////////////////
// Query MOEX on engine, market and primary board for the specified asset
// ///////////////////////////////////////////////////////////////////
func GetEngineMarketBoard(asset string) (string, string, string) {
	slog.Debug(fmt.Sprintf("Quering MOEX on engine/market for %s", asset))
	url := fmt.Sprintf("https://iss.moex.com/iss/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=boards", asset)
	assetDescription, err := Query[AssetDescription](url)
	if err != nil {
		log.Fatal("Failed to get asset description from MOEX")
	}

	info := assetDescription[1].Boards[0]
	if info.IsPrimary != 1 {
		log.Fatal("First board is not primary!")
	}

	slog.Debug(fmt.Sprintf(
		"For %s on MOEX engine is %s, market is %s, primary board is %s",
		asset,
		info.Engine,
		info.Market,
		info.Boardid,
	))
	return info.Engine, info.Market, info.Boardid
}
