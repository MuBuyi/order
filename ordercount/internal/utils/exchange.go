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

// 获取最新汇率（PHP/IDR/MYR/USD 相对 CNY），缓存10分钟
func GetRates() (map[string]float64, error) {
	ratesCache.Lock()
	defer ratesCache.Unlock()
	if ratesCache.Rates != nil && time.Since(ratesCache.Last) < 10*time.Minute {
		return ratesCache.Rates, nil
	}
	// 使用 open.er-api.com 的免费接口：基准货币为 CNY，返回完整 rates 表
	// 示例：{"result":"success","base_code":"CNY","rates":{"PHP":8.49,"IDR":2409,...}}
	url := "https://open.er-api.com/v6/latest/CNY"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Result   string             `json:"result"`
		BaseCode string             `json:"base_code"`
		Rates    map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Result != "success" {
		return nil, fmt.Errorf("exchange api returned result=%s", data.Result)
	}
	if len(data.Rates) == 0 {
		return nil, fmt.Errorf("exchange api returned empty rates")
	}
	conv := map[string]float64{}
	// 只取我们关心的几个币种，保持含义为：1 CNY ≈ conv[币种] 外币
	for _, code := range []string{"PHP", "IDR", "MYR", "USD"} {
		if v, ok := data.Rates[code]; ok {
			conv[code] = v
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
	case "USD":
		return "美元"
	case "CNY":
		return "人民币"
	}
	return cur
}