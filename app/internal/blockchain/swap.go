package blockchain

import (
	"fmt"

	"github.com/r-pine/demo_aggregation/app/internal/entity"
)

const (
	NanoUnit   = 1000000000.0
	networkFee = 0.4 * NanoUnit
)

func Swap(
	amountToFloat float64,
	aggregation entity.Aggregation,
	tonToAPine bool,
) (float64, float64, float64, float64, float64) {
	inputAmount := amountToFloat * NanoUnit

	bestOutput := 0.0
	bestCombination := ""

	var privateAmountIn, stonfiAmountIn, dedustAmountIn float64

	// Считаем вывод для каждого DEX
	fmt.Println("Обмен на одном DEX:")
	var bestName string
	var bestPortion float64
	for name, dex := range aggregation.Dex {
		portion := inputAmount - networkFee
		output := exchangeOnSingleDex(dex, inputAmount, networkFee, tonToAPine)
		percentage := (portion / inputAmount) * 100
		if tonToAPine {
			fmt.Printf("- %s: %.6f aPine, отправляем %.6f TON (%.2f%% после комиссии)\n", name, output, portion/NanoUnit, percentage)
		} else {
			fmt.Printf("- %s: %.6f TON, отправляем %.6f aPine (%.2f%% после комиссии)\n", name, output, portion/NanoUnit, percentage)
		}

		if output > bestOutput {
			bestOutput = output
			bestCombination = fmt.Sprintf("Обмен на одном DEX: %s", name)
			bestName = name
			bestPortion = portion
		}
	}

	if bestName == "private" {
		privateAmountIn = bestPortion / NanoUnit
	}
	if bestName == "stonfi" {
		stonfiAmountIn = bestPortion / NanoUnit
	}
	if bestName == "dedust" {
		dedustAmountIn = bestPortion / NanoUnit
	}

	// Считаем вывод для двух DEX
	fmt.Println("\nОбмен на двух DEX:")
	combinations := [][2]string{
		{"stonfi", "dedust"},
		{"stonfi", "private"},
		{"dedust", "private"},
	}
	var bestPortion1, bestPortion2 float64
	bestCombIndex := 10
	for i, combination := range combinations {
		dex1 := aggregation.Dex[combination[0]]
		dex2 := aggregation.Dex[combination[1]]
		output, portion1, portion2 := exchangeOnTwoDex(dex1, dex2, inputAmount, networkFee, tonToAPine)

		percentage1 := (portion1 / inputAmount) * 100
		percentage2 := (portion2 / inputAmount) * 100
		if tonToAPine {
			fmt.Printf("- %s + %s: %.6f aPine, отправляем %.6f TON (%.2f%%) + %.6f TON (%.2f%%)\n", combination[0], combination[1], output,
				portion1/NanoUnit, percentage1, portion2/NanoUnit, percentage2)
		} else {
			fmt.Printf("- %s + %s: %.6f TON, отправляем %.6f aPine (%.2f%%) + %.6f aPine (%.2f%%)\n", combination[0], combination[1], output,
				portion1/NanoUnit, percentage1, portion2/NanoUnit, percentage2)
		}

		if output > bestOutput {
			bestOutput = output
			bestCombination = fmt.Sprintf("Обмен на двух DEX: %s + %s", combination[0], combination[1])

			bestPortion1 = portion1
			bestPortion2 = portion2
			bestCombIndex = i
		}
	}
	if bestPortion1 != 0.0 && bestPortion2 != 0.0 {
		if combinations[bestCombIndex][1] == "private" {
			privateAmountIn = bestPortion2 / NanoUnit
		}
		if combinations[bestCombIndex][1] == "dedust" {
			dedustAmountIn = bestPortion2 / NanoUnit
		}
		if combinations[bestCombIndex][0] == "dedust" {
			dedustAmountIn = bestPortion1 / NanoUnit
		}
		if combinations[bestCombIndex][0] == "stonfi" {
			stonfiAmountIn = bestPortion1 / NanoUnit
		}
	}

	// Считаем вывод для всех трех DEX
	fmt.Println("\nОптимальное распределение для обмена на всех DEX:")
	bestDistribution, totalOutput := findOptimalDistribution(aggregation, inputAmount, networkFee, tonToAPine)
	for name, portion := range bestDistribution {
		percentage := (portion / inputAmount) * 100
		if tonToAPine {
			fmt.Printf("- %s: отправляем %.6f TON (%.2f%% после комиссии)\n", name, portion/NanoUnit, percentage)
		} else {
			fmt.Printf("- %s: отправляем %.6f aPine (%.2f%% после комиссии)\n", name, portion/NanoUnit, percentage)
		}
	}

	if tonToAPine {
		fmt.Printf("Общий выход: %.6f aPine\n", totalOutput)
	} else {
		fmt.Printf("Общий выход: %.6f TON\n", totalOutput)
	}

	if totalOutput > bestOutput {
		fmt.Println("totalOutput", totalOutput)
		bestOutput = totalOutput
		bestCombination = "Оптимальное распределение на всех DEX"
		for name, portion := range bestDistribution {
			if name == "private" {
				privateAmountIn = portion / NanoUnit
			}
			if name == "stonfi" {
				stonfiAmountIn = portion / NanoUnit
			}
			if name == "dedust" {
				dedustAmountIn = portion / NanoUnit
			}
		}
	}

	if tonToAPine {
		fmt.Printf("\nСамая лучшая комбинация: %s с общим выходом %.6f aPine\n", bestCombination, bestOutput)
	} else {
		fmt.Printf("\nСамая лучшая комбинация: %s с общим выходом %.6f TON\n", bestCombination, bestOutput)
	}
	return inputAmount, privateAmountIn, stonfiAmountIn, dedustAmountIn, bestOutput
}

// Функция для расчета выхода с учетом резервов и комиссий
func calculateOutput(inputAmount, reserveIn, reserveOut, dexFee, networkFee float64) float64 {
	adjustedInput := inputAmount - networkFee // Сначала вычитаем сетевую комиссию
	if adjustedInput <= 0 {
		return 0 // Недостаточно средств для покрытия сетевой комиссии
	}
	feeMultiplier := (10000 - dexFee) / 10000 // Учитываем комиссию DEX

	amountInWithFee := adjustedInput * feeMultiplier
	numerator := amountInWithFee * reserveOut
	denominator := reserveIn + amountInWithFee

	return numerator / denominator / NanoUnit // Преобразование из нано-единиц в читаемые
}

// Функция для расчета предельной цены (эффективности DEX)
func calculatePrice(reserveIn, reserveOut, dexFee, networkFee float64) float64 {
	feeMultiplier := (10000 - dexFee) / 10000 // Учитываем комиссию DEX
	return (reserveOut / (reserveIn + networkFee)) * feeMultiplier
}

// Функция для оптимального распределения средств между DEX
func findOptimalDistribution(aggregation entity.Aggregation, inputAmount, networkFee float64, tonToAPine bool) (map[string]float64, float64) {
	// Рассчитаем "эффективность" каждого DEX в зависимости от его резервов и комиссий
	priceMap := make(map[string]float64)
	totalPrice := 0.0
	for name, dex := range aggregation.Dex {
		price := calculatePrice(dex.Reserve0, dex.Reserve1, dex.Fee, networkFee)
		if !tonToAPine {
			price = calculatePrice(dex.Reserve1, dex.Reserve0, dex.Fee, networkFee)
		}
		priceMap[name] = price
		totalPrice += price
	}

	// Теперь распределим inputAmount пропорционально цене каждого DEX
	bestDistribution := make(map[string]float64)
	remainingAmount := inputAmount
	totalOutput := 0.0

	for name, price := range priceMap {
		portion := (price / totalPrice) * inputAmount

		// Вычитаем сетевую комиссию для каждой транзакции
		if portion > networkFee {
			portion -= networkFee
		} else {
			portion = 0
		}

		bestDistribution[name] = portion
		remainingAmount -= portion + networkFee

		// Рассчитываем выход для текущей порции
		if portion > 0 {
			dex := aggregation.Dex[name]
			if tonToAPine {
				totalOutput += calculateOutput(portion, dex.Reserve0, dex.Reserve1, dex.Fee, networkFee)
			} else {
				totalOutput += calculateOutput(portion, dex.Reserve1, dex.Reserve0, dex.Fee, networkFee)
			}
		}

		if remainingAmount <= 0 {
			break
		}
	}

	return bestDistribution, totalOutput
}

// Функция для расчета обмена на одном DEX с выводом суммы TON после вычета сетевой комиссии
func exchangeOnSingleDex(dex entity.Platform, inputAmount, networkFee float64, tonToAPine bool) float64 {
	portion := inputAmount - networkFee
	if !tonToAPine {
		return calculateOutput(portion, dex.Reserve1, dex.Reserve0, dex.Fee, networkFee)
	}
	return calculateOutput(portion, dex.Reserve0, dex.Reserve1, dex.Fee, networkFee)
}

// Функция для расчета обмена на двух DEX с выводом суммы TON после вычета сетевой комиссии
func exchangeOnTwoDex(dex1, dex2 entity.Platform, inputAmount, networkFee float64, tonToAPine bool) (float64, float64, float64) {
	portion := inputAmount / 2.0
	portion1 := portion - networkFee
	portion2 := portion - networkFee
	output1 := calculateOutput(portion1, dex1.Reserve0, dex1.Reserve1, dex1.Fee, networkFee)
	output2 := calculateOutput(portion2, dex2.Reserve0, dex2.Reserve1, dex2.Fee, networkFee)
	if !tonToAPine {
		output1 = calculateOutput(portion1, dex1.Reserve1, dex1.Reserve0, dex1.Fee, networkFee) // Поменяли резервы местами
		output2 = calculateOutput(portion2, dex2.Reserve1, dex2.Reserve0, dex2.Fee, networkFee) // Поменяли резервы местами
	}
	return output1 + output2, portion1, portion2
}
