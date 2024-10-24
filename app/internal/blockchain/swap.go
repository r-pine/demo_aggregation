package blockchain

import (
	"fmt"

	"math/rand"

	"github.com/r-pine/demo_aggregation/app/internal/entity"
)

const (
	netComs   = 0.12 * 1_000_000_000
	precision = 1000
)

func Swap(
	amountToFloat float64,
	aggregation entity.Aggregation,
	tonToAPine bool,
) (float64, float64, float64, float64) {
	inputAmount := amountToFloat * NanoUnit

	var bestX1, bestX2, bestX3 float64

	bestReward := 0.0
	bestSwapValue := 0.0

	for i := 0; i < 2000000; i++ {
		pX1 := rand.Intn(precision)
		x1 := float64(pX1) * amountToFloat / precision
		if pX1 < 3 {
			x1 = 0
			pX1 = 0
		}

		pX2 := rand.Intn(precision - pX1)

		x2 := float64(pX2) * amountToFloat / precision
		if pX2 < 3 {
			x2 = 0
			pX2 = 0
		}

		pX3 := precision - pX1 - pX2

		x3 := float64(pX3) * amountToFloat / precision
		if pX3 < 3 {
			if x2 == 0 {
				x1 += x3
			} else {
				x2 += x3
			}
			x3 = 0
		}

		numUsedPools := 0
		totalReward := 0.0

		if tonToAPine {
			if x1 > 0 {
				dy := calculateDy(x1, aggregation.Dex["stonfi"].Reserve0, aggregation.Dex["stonfi"].Reserve1, aggregation.Dex["stonfi"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x2 > 0 {
				dy := calculateDy(x2, aggregation.Dex["dedust"].Reserve0, aggregation.Dex["dedust"].Reserve1, aggregation.Dex["dedust"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x3 > 0 {
				dy := calculateDy(x3, aggregation.Dex["private"].Reserve0, aggregation.Dex["private"].Reserve1, aggregation.Dex["private"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if getTotalSwapValue(tonToAPine, amountToFloat, totalReward, float64(netComs*numUsedPools)) > bestSwapValue {

				bestReward = totalReward
				bestSwapValue = getTotalSwapValue(tonToAPine, amountToFloat, totalReward, float64(netComs*numUsedPools))
				bestX1 = x1
				bestX2 = x2
				bestX3 = x3
			}
		} else {
			if x1 > 0 {
				dy := calculateDy(x1, aggregation.Dex["stonfi"].Reserve1, aggregation.Dex["stonfi"].Reserve0, aggregation.Dex["stonfi"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x2 > 0 {
				dy := calculateDy(x2, aggregation.Dex["dedust"].Reserve1, aggregation.Dex["dedust"].Reserve0, aggregation.Dex["dedust"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x3 > 0 {
				dy := calculateDy(x3, aggregation.Dex["private"].Reserve1, aggregation.Dex["private"].Reserve0, aggregation.Dex["private"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if getTotalSwapValue(tonToAPine, amountToFloat, totalReward, float64(netComs*numUsedPools)) > bestSwapValue {
				bestReward = totalReward
				bestSwapValue = getTotalSwapValue(tonToAPine, amountToFloat, totalReward, float64(netComs*numUsedPools))
				bestX1 = x1
				bestX2 = x2
				bestX3 = x3
			}

		}
	}

	fmt.Printf("Лучшее полученное вознаграждение: %.2f\n", bestReward/1_000_000_000)
	fmt.Printf("Значения x1, x2, x3 для лучшего варианта: %.2f, %.2f, %.2f\n",
		bestX1/1_000_000_000, bestX2/1_000_000_000, bestX3/1_000_000_000)

	fmt.Printf("Значения x1, x2, x3 для лучшего варианта: %.2f, %.2f, %.2f\n",
		((bestX1 / inputAmount) * 100), (bestX2/inputAmount)*100, (bestX3/inputAmount)*100)

	return bestX1, bestX2, bestX3, bestReward / 1_000_000_000
}

func calculateDy(dx, reserveIn, reserveOut float64, fee float64) float64 {
	feeFraction := fee / 10000.0
	dxWithFee := dx * (1 - feeFraction)
	return (dxWithFee * reserveOut) / (reserveIn + dxWithFee)
}

func getTotalSwapValue(tonToAPine bool, sumIn, sumOut, net_comission float64) float64 {
	res := 0.0
	if tonToAPine {
		res = sumOut / (sumIn + net_comission)
	} else {
		res = (sumOut - net_comission) / sumIn
	}
	return res
}
