package manual

import "math"

// ComputeQuota 把人民币金额按当前汇率换算成 quota。
//
//	amountYuan  = amountMoney / 10000.0
//	amountUSD   = amountYuan / exchangeRateCNYPerUSD
//	amountQuota = floor(amountUSD * baseQuotaPerDollar)
//
// 全部用 float64 做中间运算,最后一步 floor 成 int64,极小浮点误差不会让用户多收 / 少收。
//
// 异常情况(exchangeRate <= 0 或 baseQuotaPerDollar <= 0)返回 0;
// 调用方应当在更上层校验配置并报错(handler / service 兜底)。
func ComputeQuota(amountMoney int64, exchangeRateCNYPerUSD float64, baseQuotaPerDollar int64) int64 {
	if amountMoney <= 0 || exchangeRateCNYPerUSD <= 0 || baseQuotaPerDollar <= 0 {
		return 0
	}
	yuan := float64(amountMoney) / 10000.0
	usd := yuan / exchangeRateCNYPerUSD
	return int64(math.Floor(usd * float64(baseQuotaPerDollar)))
}
