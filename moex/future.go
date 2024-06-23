package moex

import (
	"fmt"
	"log"
	"log/slog"
	"regexp"
)

// ///////////////////////////////////////////////////////////////////
// Query MOEX on future's underlying asset code
// ///////////////////////////////////////////////////////////////////
func GetFutureUnderlyingAsset(future string) string {
	slog.Debug(fmt.Sprintf("Quering MOEX on %s future", future))

	url := fmt.Sprintf("https://iss.moex.com/iss/engines/futures/markets/forts/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=securities",
		future)
	assetInfo, err := Query[AssetInfo](url)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to query MOEX: %w", err))
	}

	asset := assetInfo[1].Securities[0].Assetcode
	// For GLDRUBF future MOEX returns GLDRUBTOM instead of GLDRUB_TOM
	var re = regexp.MustCompile("(TOM)$")
	return re.ReplaceAllString(asset, "_TOM")
}
