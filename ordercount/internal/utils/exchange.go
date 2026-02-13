package utils

import (
	"encoding/json"
	"fmt"
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
	// 用 /live endpoint，source=CNY
	url := "http://api.exchangerate.host/live?access_key=b602a5c72c93c34b68cc4aef9259f29e&source=CNY&currencies=PHP,IDR,MYR&format=1"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Quotes map[string]float64 `json:"quotes"`
		Success bool `json:"success"`
		Error   interface{} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if !data.Success {
		return nil, fmt.Errorf("exchange api error: %v", data.Error)
	}
	conv := map[string]float64{}
	// 解析如 "CNYPHP": 7.8
	for k, v := range data.Quotes {
		if len(k) == 6 && k[:3] == "CNY" {
			conv[k[3:]] = v
		}
	}
	conv["CNY"] = 1
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