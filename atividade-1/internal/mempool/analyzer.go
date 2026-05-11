package mempool

import (
	"math"

	"bitcoin-monitor-t1/internal/rpc"
)



const (
	ThresholdLowMedium  = 10.0 
	ThresholdMediumHigh = 50.0 
)


type FeeDistribution struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
}


type Summary struct {
	TxCount         int64           `json:"tx_count"`
	TotalVsize      int64           `json:"total_vsize"`
	AvgFeeRate      float64         `json:"avg_fee_rate"`
	MinFeeRate      float64         `json:"min_fee_rate"`
	MaxFeeRate      float64         `json:"max_fee_rate"`
	FeeDistribution FeeDistribution `json:"fee_distribution"`
}




func Analyze(entries map[string]rpc.RawMempoolEntry) Summary {
	if len(entries) == 0 {
		return Summary{}
	}

	var (
		totalVsize  int64
		totalFeeRate float64
		minRate     = math.MaxFloat64
		maxRate     = -math.MaxFloat64
		dist        FeeDistribution
	)

	for _, e := range entries {
		if e.VSize <= 0 {
			continue
		}

		
		feeSat := e.Fee * 1e8
		rate := feeSat / float64(e.VSize)

		totalVsize += e.VSize
		totalFeeRate += rate

		if rate < minRate {
			minRate = rate
		}
		if rate > maxRate {
			maxRate = rate
		}

		switch classify(rate) {
		case "low":
			dist.Low++
		case "medium":
			dist.Medium++
		case "high":
			dist.High++
		}
	}

	count := int64(len(entries))
	avg := 0.0
	if count > 0 {
		avg = totalFeeRate / float64(count)
	}

	if minRate == math.MaxFloat64 {
		minRate = 0
	}
	if maxRate == -math.MaxFloat64 {
		maxRate = 0
	}

	return Summary{
		TxCount:         count,
		TotalVsize:      totalVsize,
		AvgFeeRate:      round2(avg),
		MinFeeRate:      round2(minRate),
		MaxFeeRate:      round2(maxRate),
		FeeDistribution: dist,
	}
}


func classify(rateSatPerVbyte float64) string {
	switch {
	case rateSatPerVbyte < ThresholdLowMedium:
		return "low"
	case rateSatPerVbyte < ThresholdMediumHigh:
		return "medium"
	default:
		return "high"
	}
}


func Classify(rateSatPerVbyte float64) string {
	return classify(rateSatPerVbyte)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}