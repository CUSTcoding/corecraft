package services

import (
	"math"
	"corecraft-atividade-1/rpc"
	"corecraft-atividade-1/models"
)

func ClassifyFee(rate float64) string {
	if rate < 10 {
		return "low"
	} else if rate <= 50 {
		return "medium"
	}
	return "high"
}


func cleanFloat(f float64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return f
}

func GetMempoolSummary(client *rpc.RPCClient) (*models.MempoolSummary, error) {

	resp, err := client.Call("getrawmempool", []interface{}{true})
	if err != nil {
		return nil, err
	}

	result := resp["result"].(map[string]interface{})

	var totalFeeRate float64
	var totalVSize float64
	minFee := 0.0
	maxFee := 0.0
	first := true

	distribution := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	count := 0


	for _, tx := range result {
		txData := tx.(map[string]interface{})

		fee := txData["fees"].(map[string]interface{})["base"].(float64)
		vsize := txData["vsize"].(float64)

		if vsize == 0 {
			continue
		}

		feeRate := fee / vsize

		if math.IsNaN(feeRate) || math.IsInf(feeRate, 0) {
			continue
		}

		totalFeeRate += feeRate
		totalVSize += vsize

		if first {
			minFee = feeRate
			maxFee = feeRate
			first = false
		} else {
			if feeRate < minFee {
				minFee = feeRate
			}
			if feeRate > maxFee {
				maxFee = feeRate
			}
		}

		category := ClassifyFee(feeRate)
		distribution[category]++

		count++
	}

	avgFee := 0.0
	if count > 0 {
		avgFee = totalFeeRate / float64(count)
	}

	return &models.MempoolSummary{
		TxCount:         count,
		TotalVSize:      totalVSize,
		AvgFeeRate:      cleanFloat(avgFee),
		MinFeeRate:      cleanFloat(minFee),
		MaxFeeRate:      cleanFloat(maxFee),
		FeeDistribution: distribution,
	}, nil
}