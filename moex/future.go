package moex

import (
	"fmt"
	"log/slog"
	"regexp"
)

type assetInfo []struct {
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

// ///////////////////////////////////////////////////////////////////
// Query MOEX on future's underlying asset code
// ///////////////////////////////////////////////////////////////////
func (asset *Asset) GetFutureUnderlyingAsset() (Asset, error) {
	slog.Debug(fmt.Sprintf("Quering MOEX on %s future", asset.Secid))

	url := fmt.Sprintf("https://iss.moex.com/iss/engines/futures/markets/forts/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=securities",
		asset.Secid)
	assetInfo, err := query[assetInfo](url)
	if err != nil {
		return Asset{}, err
	}

	// For GLDRUBF future MOEX returns GLDRUBTOM instead of GLDRUB_TOM
	var re = regexp.MustCompile("(TOM)$")
	var baseAssetCode = re.ReplaceAllString(assetInfo[1].Securities[0].Assetcode, "_TOM")

	return GetAsset(baseAssetCode)
}
