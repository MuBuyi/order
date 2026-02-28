package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	exch "ordercount/internal/utils"
)

// 对外汇率接口：返回主要币种相对人民币的汇率（1 单位外币 = ? 人民币）
// GET /api/exchange/rates
func ExchangeRates(c *gin.Context) {
	rates, err := exch.GetRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 补充中文名称，方便前端展示
	labels := map[string]string{
		"PHP": exch.CurrencyName("PHP"),
		"IDR": exch.CurrencyName("IDR"),
		"MYR": exch.CurrencyName("MYR"),
		"USD": exch.CurrencyName("USD"),
		"CNY": exch.CurrencyName("CNY"),
	}

	c.JSON(http.StatusOK, gin.H{
		"rates":  rates,
		"labels": labels,
	})
}
