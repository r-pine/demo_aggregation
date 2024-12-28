package main

import (
	"fmt"
	"math/rand"
)

const (
	tonToaPine = "TON_TO_APINE"
	aPineToTon = "APINE_TO_TON"
)

type Message struct {
	AmountTon        string `json:"amount_ton"`
	DstAddress       string `json:"dst_address"`
	Payload          string `json:"payload"`
	CalculatedAmount string `json:"calculated_amount"`
}

type BodyResponse struct {
	Msgs                []Message `json:"messages"`
	SumCalculatedAmount string    `json:"sum_calculated_amount"`
}

type Aggregation struct {
	Dex map[string]Platform `json:"dex"`
}

type Platform struct {
	Name     string  `json:"name"`
	Address  Address `json:"address"`
	Fee      int     `json:"fee"`
	Reserve0 float64 `json:"reserve0"`
	Reserve1 float64 `json:"reserve1"`
	IsActive bool    `json:"is_active"`
	Status   string  `json:"status"`
	Balance  string  `json:"balance"`
	Price    float64 `json:"-"`
	NewPrice float64 `json:"-"`
	Dx       float64 `json:"-"`
	Dy       float64 `json:"-"`
}

type Address struct {
	Bounce   string `json:"bounce"`
	UnBounce string `json:"unbounce"`
}

const (
	totalAmount = 10.0 * 1_000_000_000
	direction   = 1
	netComs     = 0.03 * 1_000_000_000
	precision   = 1000
)

func calculateDy(dx, reserveIn, reserveOut float64, fee int) float64 {
	feeFraction := float64(fee) / 10000.0
	dxWithFee := dx * (1 - feeFraction)
	return (dxWithFee * reserveOut) / (reserveIn + dxWithFee)
}

func getTotalSwapValue(dir int, sumIn, sumOut, net_comission float64) float64 {
	res := 0.0
	if dir == 1 {
		res = sumOut / (sumIn + net_comission)
	}
	if dir == 0 {
		res = (sumOut - net_comission) / sumIn
	}
	return res
}

func main() {
	var bestX1, bestX2, bestX3 float64

	baseReserves := Aggregation{
		Dex: map[string]Platform{
			"stonfi": {
				Name:     "stonfi",
				Reserve0: 101935334773,
				Reserve1: 3273580141059169,
				Fee:      30,
			},
			"dedust": {
				Name:     "dedust",
				Reserve0: 100890720033,
				Reserve1: 3241290194034248,
				Fee:      25,
			},
			"private": {
				Name:     "private",
				Reserve0: 101946149862,
				Reserve1: 3273338639122545,
				Fee:      30,
			},
		},
	}

	bestReward := 0.0
	bestSwapValue := 0.0

	for i := 0; i < 2000000; i++ { // Количество итераций
		pX1 := rand.Intn(precision)
		x1 := float64(pX1) * totalAmount / precision
		if pX1 < 3 {
			x1 = 0
			pX1 = 0
		}

		pX2 := rand.Intn(precision - pX1)

		x2 := float64(pX2) * totalAmount / precision
		if pX2 < 3 {
			x2 = 0
			pX2 = 0
		}

		pX3 := precision - pX1 - pX2

		x3 := float64(pX3) * totalAmount / precision
		if pX3 < 3 {
			if x2 == 0 {
				x1 += x3
			} else {
				x2 += x3
			}
			x3 = 0
		}

		// Сброс начальных резерваций для каждого пула
		res := baseReserves

		numUsedPools := 0
		totalReward := 0.0

		// ton to jetton
		if direction == 1 {
			// Обработка первого пула (stonfi)
			if x1 > 0 {
				dy := calculateDy(x1, res.Dex["stonfi"].Reserve0, res.Dex["stonfi"].Reserve1, res.Dex["stonfi"].Fee)
				totalReward += dy
				numUsedPools++
			}

			// Обработка второго пула (dedust)
			if x2 > 0 {
				dy := calculateDy(x2, res.Dex["dedust"].Reserve0, res.Dex["dedust"].Reserve1, res.Dex["dedust"].Fee)
				totalReward += dy
				numUsedPools++
			}

			// Обработка третьего пула (private)
			if x3 > 0 {
				dy := calculateDy(x3, res.Dex["private"].Reserve0, res.Dex["private"].Reserve1, res.Dex["private"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if getTotalSwapValue(direction, totalAmount, totalReward, float64(netComs*numUsedPools)) > bestSwapValue {

				bestReward = totalReward
				bestSwapValue = getTotalSwapValue(direction, totalAmount, totalReward, float64(netComs*numUsedPools))
				bestX1 = x1
				bestX2 = x2
				bestX3 = x3
			}
		}
		if direction == 0 {
			if x1 > 0 {
				dy := calculateDy(x1, res.Dex["stonfi"].Reserve1, res.Dex["stonfi"].Reserve0, res.Dex["stonfi"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x2 > 0 {
				dy := calculateDy(x2, res.Dex["dedust"].Reserve1, res.Dex["dedust"].Reserve0, res.Dex["dedust"].Fee)
				totalReward += dy
				numUsedPools++
			}

			if x3 > 0 {
				dy := calculateDy(x3, res.Dex["private"].Reserve1, res.Dex["private"].Reserve0, res.Dex["private"].Fee)
				totalReward += dy
				numUsedPools++
			}

			//if totalReward > bestReward
			if getTotalSwapValue(direction, totalAmount, totalReward, float64(netComs*numUsedPools)) > bestSwapValue {
				bestReward = totalReward
				bestSwapValue = getTotalSwapValue(direction, totalAmount, totalReward, float64(netComs*numUsedPools))
				bestX1 = x1
				bestX2 = x2
				bestX3 = x3
			}

		}
	}

	fmt.Printf("Лучшее полученное вознаграждение: %.2f\n", bestReward/1_000_000_000)
	fmt.Printf("Значения x1, x2, x3 для лучшего варианта: %.2f, %.2f, %.2f\n", bestX1/1_000_000_000, bestX2/1_000_000_000, bestX3/1_000_000_000)
	fmt.Println(bestX1/1_000_000_000 + bestX2/1_000_000_000 + bestX3/1_000_000_000)
}
