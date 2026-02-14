package models

import "time"

// DailySettlement 记录每日的结账/利润明细
// 一天可以有多次结算记录，按 date + created_at 区分
type DailySettlement struct {
    ID uint `gorm:"primaryKey" json:"id"`

    // 结算日期（只保留到天，统一用本地日期字符串）
    Date string `json:"date" gorm:"size:10;index"`

    Country  string  `json:"country" gorm:"size:20"`
    Currency string  `json:"currency" gorm:"size:10"`

    // 输入的原始数值（以外币为单位）
    SaleTotal  float64 `json:"sale_total"`   // 当天销售总额
    AdCost     float64 `json:"ad_cost"`      // 广告费（原始输入）
    Exchange   float64 `json:"exchange"`     // 使用的汇率（1 外币 ≈ ? 人民币）
    GoodsCost  float64 `json:"goods_cost"`   // 货款成本
    ShuaDanFee float64 `json:"shua_dan_fee"` // 刷单费用
    FixedCost  float64 `json:"fixed_cost"`   // 固定成本

    // 中间计算结果（人民币）
    AdDeduction   float64 `json:"ad_deduction"`   // 广告成本折算：(广告费 + 广告费*11%) * 汇率
    PlatformFee   float64 `json:"platform_fee"`   // 平台手续费：当天销售总额 * 7%（默认同币种，是否折算看业务）
    Profit        float64 `json:"profit"`         // 最终利润

    // 备注信息，例如活动说明、特殊情况等
    Remark string `json:"remark" gorm:"size:255"`

    CreatedAt time.Time `json:"created_at"`
}
