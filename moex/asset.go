package moex

import (
	"fmt"
	"log"
	"log/slog"
)

type Asset struct {
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
}

type AssetDescription []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Boards []Asset `json:"boards,omitempty"`
}

// ///////////////////////////////////////////////////////////////////
// Query MOEX on engine, market and primary board for the specified asset
// ///////////////////////////////////////////////////////////////////
func GetAsset(asset string) (Asset, error) {
	slog.Debug(fmt.Sprintf("Quering MOEX on engine/market for %s", asset))
	url := fmt.Sprintf("https://iss.moex.com/iss/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=boards", asset)
	assetDescription, err := query[AssetDescription](url)
	if err != nil {
		return Asset{}, err
	}

	if len(assetDescription) == 0 || len(assetDescription[1].Boards) == 0 {
		return Asset{}, fmt.Errorf("asset %s not found on MOEX", asset)
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
	return info, nil
}
