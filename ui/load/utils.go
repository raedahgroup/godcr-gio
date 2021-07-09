package load

import (
	"encoding/json"
	"net/http"

	"gioui.org/widget"
)

func fetchExchangeValue(target interface{}) {
	url := "https://api.bittrex.com/v3/markets/DCR-USDT/ticker"
	res, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		log.Error(err)
	}
}

func mustIcon(ic *widget.Icon, err error) *widget.Icon {
	if err != nil {
		panic(err)
	}
	return ic
}
