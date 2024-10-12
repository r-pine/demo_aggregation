package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
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

func (c *Controller) GetSwapPayload(ctx *gin.Context) {
	type bodyRequest struct {
		Amount    string `json:"amount"`
		Address   string `json:"address"`
		Direction string `json:"direction"`
	}
	var br bodyRequest

	if err := ctx.ShouldBind(&br); err != nil {
		ctx.JSON(http.StatusBadRequest, c.getErrorResponse())
		return
	}

	data, err := c.sc.Get("states")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var res *entity.Aggregation
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Выбираем направление обмена: true для TON->aPine, false для aPine->TON
	swapTonToApine := true

	var dexes []entity.Platform
	dexesMap := res.Dex
	for i := range dexesMap {
		a := dexesMap[i]
		dexes = append(dexes, a)
	}

	var totalInput float64
	if swapTonToApine {
		totalInput = 3_000_000_000 // 3 TON в нанотонах
	} else {
		totalInput = 100_000_000_000 // Например, 100 aPine в наноединицах
	}

	remainingInput := totalInput
	totalOutput := 0.0

	// Вычисляем цены на каждом DEX
	for i := range dexes {
		dexes[i].Price = calculateNewPrice(dexes[i].Reserve0, dexes[i].Reserve1)
	}

	// Сортируем DEX по лучшей цене с учетом комиссии
	sort.Slice(dexes, func(i, j int) bool {
		var effectivePriceI, effectivePriceJ float64
		feeI := float64(dexes[i].PoolFee) / 10000.0
		feeJ := float64(dexes[j].PoolFee) / 10000.0

		if swapTonToApine {
			// Для TON -> aPine, учитываем уменьшение dx из-за комиссии
			effectivePriceI = (dexes[i].Reserve1 / dexes[i].Reserve0) * (1 - feeI)
			effectivePriceJ = (dexes[j].Reserve1 / dexes[j].Reserve0) * (1 - feeJ)
			return effectivePriceI > effectivePriceJ
		} else {
			// Для aPine -> TON, учитываем уменьшение dy из-за комиссии
			effectivePriceI = (dexes[i].Reserve0 / dexes[i].Reserve1) * (1 - feeI)
			effectivePriceJ = (dexes[j].Reserve0 / dexes[j].Reserve1) * (1 - feeJ)
			return effectivePriceI < effectivePriceJ
		}
	})

	// Распределяем токены между DEX
	for i := 0; i < len(dexes); i++ {
		if remainingInput <= 0 {
			break
		}

		var dx, dy float64

		if swapTonToApine {
			// Обмен TON на aPine
			dx = remainingInput // Пытаемся обменять весь оставшийся TON
			dy = calculateDy(dx, dexes[i].Reserve0, dexes[i].Reserve1, dexes[i].PoolFee)
		} else {
			// Обмен aPine на TON
			dx = remainingInput // Пытаемся обменять весь оставшийся aPine
			dy = calculateDy(dx, dexes[i].Reserve1, dexes[i].Reserve0, dexes[i].PoolFee)
		}

		// Обновляем резервы после обмена
		var newReserveIn, newReserveOut float64
		if swapTonToApine {
			newReserveIn = dexes[i].Reserve0 + dx
			newReserveOut = dexes[i].Reserve1 - dy
		} else {
			newReserveIn = dexes[i].Reserve0 - dy
			newReserveOut = dexes[i].Reserve1 + dx
		}

		newPrice := calculateNewPrice(newReserveIn, newReserveOut)

		// Проверяем, не стала ли цена хуже, чем на следующем DEX
		if i < len(dexes)-1 {
			var nextEffectivePrice float64
			nextFee := float64(dexes[i+1].PoolFee) / 10000.0

			if swapTonToApine {
				nextEffectivePrice = (dexes[i+1].Reserve1 / dexes[i+1].Reserve0) * (1 - nextFee)
				if newPrice*(1-float64(dexes[i].PoolFee)/10000.0) < nextEffectivePrice {
					continue
				}
			} else {
				nextEffectivePrice = (dexes[i+1].Reserve0 / dexes[i+1].Reserve1) * (1 - nextFee)
				if newPrice*(1-float64(dexes[i].PoolFee)/10000.0) > nextEffectivePrice {
					continue
				}
			}
		}

		// Обновляем данные DEX
		dexes[i].Dx = dx
		dexes[i].Dy = dy
		dexes[i].NewPrice = newPrice
		dexes[i].Reserve0 = newReserveIn
		dexes[i].Reserve1 = newReserveOut

		totalOutput += dy
		remainingInput -= dx
	}

	// Выводим результаты
	if swapTonToApine {
		fmt.Printf("Общее количество полученного aPine: %.6f\n", totalOutput/1e9)
		fmt.Println("Распределение TON между DEX:")
	} else {
		fmt.Printf("Общее количество полученного TON: %.6f\n", totalOutput/1e9)
		fmt.Println("Распределение aPine между DEX:")
	}

	if remainingInput > 0 {
		if swapTonToApine {
			fmt.Printf("Осталось неиспользованных TON: %.6f\n", remainingInput/1e9)
		} else {
			fmt.Printf("Осталось неиспользованных aPine: %.6f\n", remainingInput/1e9)
		}
	}
}

func calculateDy(dx, reserveIn, reserveOut float64, fee int) float64 {
	// Учитываем комиссию пула
	feeFraction := float64(fee) / 10000.0
	dxWithFee := dx * (1 - feeFraction)
	return (dxWithFee * reserveOut) / (reserveIn + dxWithFee)
}

func calculateDx(dy, reserveIn, reserveOut float64, fee int) float64 {
	// Учитываем комиссию пула
	feeFraction := float64(fee) / 10000.0
	numerator := reserveIn * dy
	denominator := (reserveOut - dy) * (1 - feeFraction)
	return numerator/denominator + 1e-9 // Добавляем небольшую величину для предотвращения деления на ноль
}

func calculateNewPrice(reserveIn, reserveOut float64) float64 {
	return reserveOut / reserveIn
}

func getDexPrice(res0, res1 int64) int64 {
	var dexPrice int64
	if res0 > res1 {
		dexPrice = res0 / res1
	} else {
		dexPrice = res1 / res0
	}
	return dexPrice
}

func (c *Controller) getErrorResponse() *BodyResponse {
	var bodyResponse BodyResponse
	bodyResponse.Msgs = []Message{}
	bodyResponse.SumCalculatedAmount = "0"
	return &bodyResponse
}
