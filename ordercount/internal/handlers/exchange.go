package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	ratesCache struct {
		sync.Mutex
		Rates map[string]float64
		Last time.Time
	}
)

// 获取最新汇率（PHP/IDR/MYR->CNY），缓存10分钟
func GetRates() (map[string]float64, error) {
	ratesCache.Lock()
	defer ratesCache.Unlock()
	if ratesCache.Rates != nil && time.Since(ratesCache.Last) < 10*time.Minute {
		return ratesCache.Rates, nil
	}
	resp, err := http.Get("https://api.exchangerate.host/latest?base=CNY&symbols=PHP,IDR,MYR")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	// 汇率为 1 CNY = ? 目标币，需取倒数
	conv := map[string]float64{}
	for k, v := range data.Rates {
		if v > 0 {
			conv[k] = 1 / v
		}
	}
	// CNY->CNY
	conv["CNY"] = 1
	// 缓存
	ratesCache.Rates = conv
	ratesCache.Last = time.Now()
	return conv, nil
}

// 获取币种对应的中文
func CurrencyName(cur string) string {
	switch cur {
	case "PHP":
		return "菲律宾比索"
	case "IDR":
		return "印尼盾"
	case "MYR":
		return "马来西亚林吉特"
	case "CNY":
		return "人民币"
	}
	return cur
}
