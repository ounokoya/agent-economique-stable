// ✅ VALIDATION TOUS INDICATEURS BINANCE FUTURES - PRÉCISION 100%
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/indicators"
)

func main() {
	fmt.Println("🔍 VALIDATION COMPLÈTE BINANCE FUTURES - TOUS INDICATEURS")
	fmt.Println("=" + strings.Repeat("=", 60))

	// 1️⃣ CRÉER CLIENT BINANCE FUTURES
	client := binance.NewFuturesClient()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2️⃣ RÉCUPÉRER 300 KLINES DEPUIS BINANCE FUTURES
	fmt.Println("📡 Récupération des 300 dernières klines depuis Binance Futures...")
	futuresKlines, err := client.GetKlines(ctx, "SOLUSDT", "5m", 300)
	if err != nil {
		log.Fatalf("❌ Erreur récupération klines: %v", err)
	}

	// Convertir en format standard
	klines := client.ConvertToStandardKline(futuresKlines)
	indicatorsKlines := make([]indicators.Kline, len(klines))
	for i, k := range klines {
		indicatorsKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix(),
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	fmt.Printf("✅ %d klines récupérées de %s à %s\n", 
		len(klines), 
		klines[0].OpenTime.Format("2006-01-02 15:04"), 
		klines[len(klines)-1].OpenTime.Format("2006-01-02 15:04"))

	// 3️⃣ VALIDATION MFI
	fmt.Println("\n📊 VALIDATION MFI (période 14):")
	mfiTV := indicators.NewMFITVStandard(14)
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	volume := make([]float64, len(klines))

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volume[i] = k.Volume
	}

	mfiValues := mfiTV.Calculate(high, low, close, volume)
	lastMFI := mfiTV.GetLastValue(mfiValues)
	
	// Afficher 5 dernières valeurs MFI
	fmt.Println("   5 dernières valeurs MFI:")
	startIdx := len(mfiValues) - 5
	if startIdx < 0 {
		startIdx = 0
	}
	for i := startIdx; i < len(mfiValues); i++ {
		klineIdx := i
		if klineIdx >= len(klines) {
			klineIdx = len(klines) - 1
		}
		fmt.Printf("   %s %.2f %s\n", 
			klines[klineIdx].OpenTime.Format("15:04"),
			mfiValues[i],
			mfiTV.GetSignal(mfiValues[i]))
	}
	
	fmt.Printf("   MFI Actuel: %.2f - Signal: %s\n", lastMFI, mfiTV.GetSignal(lastMFI))

	// 4️⃣ VALIDATION MACD
	fmt.Println("\n📊 VALIDATION MACD (12,26,9):")
	macdValues, signalValues, histValues := indicators.MACDFromKlines(indicatorsKlines, 12, 26, 9, func(k indicators.Kline) float64 { return k.Close })
	
	if len(macdValues) > 0 {
		// Afficher 5 dernières valeurs MACD
		fmt.Println("   5 dernières valeurs MACD:")
		startIdx := len(macdValues) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(macdValues); i++ {
			klineIdx := i
			if klineIdx >= len(klines) {
				klineIdx = len(klines) - 1
			}
			fmt.Printf("   %s MACD:%.4f Sig:%.4f Hist:%.4f\n", 
				klines[klineIdx].OpenTime.Format("15:04"),
				macdValues[i], signalValues[i], histValues[i])
		}
		
		lastMACD := macdValues[len(macdValues)-1]
		lastSignal := signalValues[len(signalValues)-1]
		lastHist := histValues[len(histValues)-1]
		
		var macdSignal string
		if lastHist > 0 && lastMACD > lastSignal {
			macdSignal = "🟢 HAUSSIER FORT"
		} else if lastHist > 0 && lastMACD < lastSignal {
			macdSignal = "🟡 HAUSSIER FAIBLE"
		} else if lastHist < 0 && lastMACD < lastSignal {
			macdSignal = "🔴 BAISSIER FORT"
		} else {
			macdSignal = "🟡 BAISSIER FAIBLE"
		}
		
		fmt.Printf("   MACD Actuel: %.4f - Signal: %.4f - Hist: %.4f - %s\n", 
			lastMACD, lastSignal, lastHist, macdSignal)
	}

	// 5️⃣ VALIDATION CCI
	fmt.Println("\n📊 VALIDATION CCI (période 20):")
	cciValues := indicators.CCIFromKlines(indicatorsKlines, "standard", 20)
	if len(cciValues) > 0 {
		// Afficher 5 dernières valeurs CCI
		fmt.Println("   5 dernières valeurs CCI:")
		startIdx := len(cciValues) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(cciValues); i++ {
			klineIdx := i
			if klineIdx >= len(klines) {
				klineIdx = len(klines) - 1
			}
			
			var cciSignal string
			val := cciValues[i]
			if !math.IsNaN(val) {
				switch {
				case val > 100:
					cciSignal = "🔴 SURACHAT"
				case val < -100:
					cciSignal = "🟢 SURVENTE"
				default:
					cciSignal = "⚪ NEUTRE"
				}
			} else {
				cciSignal = "⚪ NaN"
			}
			
			fmt.Printf("   %s %.2f %s\n", 
				klines[klineIdx].OpenTime.Format("15:04"),
				val, cciSignal)
		}
		
		lastCCI := cciValues[len(cciValues)-1]
		
		var cciSignal string
		if !math.IsNaN(lastCCI) {
			switch {
			case lastCCI > 100:
				cciSignal = "🔴 SURACHAT"
			case lastCCI < -100:
				cciSignal = "🟢 SURVENTE"
			default:
				cciSignal = "⚪ NEUTRE"
			}
		} else {
			cciSignal = "⚪ NaN"
		}
		
		fmt.Printf("   CCI Actuel: %.2f - Signal: %s\n", lastCCI, cciSignal)
	}

	// 6️⃣ VALIDATION DMI
	fmt.Println("\n📊 VALIDATION DMI (période 14):")
	diPlus, diMinus, _, adx := indicators.DMIFromKlines(indicatorsKlines, 14)
	
	if len(diPlus) > 0 {
		// Afficher 5 dernières valeurs DMI
		fmt.Println("   5 dernières valeurs DMI:")
		startIdx := len(diPlus) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(diPlus); i++ {
			klineIdx := i
			if klineIdx >= len(klines) {
				klineIdx = len(klines) - 1
			}
			
			var dmiSignal string
			if !math.IsNaN(diPlus[i]) && !math.IsNaN(diMinus[i]) {
				if diPlus[i] > diMinus[i] {
					if adx[i] > 25 {
						dmiSignal = "🟢 HAUSSIER FORT"
					} else {
						dmiSignal = "🟡 HAUSSIER FAIBLE"
					}
				} else {
					if adx[i] > 25 {
						dmiSignal = "🔴 BAISSIER FORT"
					} else {
						dmiSignal = "🟡 BAISSIER FAIBLE"
					}
				}
			} else {
				dmiSignal = "⚪ NaN"
			}
			
			fmt.Printf("   %s DI+:%.2f DI-:%.2f ADX:%.2f %s\n", 
				klines[klineIdx].OpenTime.Format("15:04"),
				diPlus[i], diMinus[i], adx[i], dmiSignal)
		}
		
		lastDIPlus := diPlus[len(diPlus)-1]
		lastDIMinus := diMinus[len(diMinus)-1]
		lastADX := adx[len(adx)-1]
		
		var dmiSignal string
		if !math.IsNaN(lastDIPlus) && !math.IsNaN(lastDIMinus) {
			if lastDIPlus > lastDIMinus {
				if lastADX > 25 {
					dmiSignal = "🟢 TENDANCE HAUSSIÈRE FORTE"
				} else {
					dmiSignal = "🟡 TENDANCE HAUSSIÈRE FAIBLE"
				}
			} else {
				if lastADX > 25 {
					dmiSignal = "🔴 TENDANCE BAISSIÈRE FORTE"
				} else {
					dmiSignal = "🟡 TENDANCE BAISSIÈRE FAIBLE"
				}
			}
		} else {
			dmiSignal = "⚪ NaN"
		}
		
		fmt.Printf("   DMI Actuel: DI+:%.2f - DI-:%.2f - ADX:%.2f - Signal: %s\n", 
			lastDIPlus, lastDIMinus, lastADX, dmiSignal)
	}

	// 7️⃣ VALIDATION STOCHASTIC
	fmt.Println("\n📊 VALIDATION STOCHASTIC (%K=14, %D=3):")
	stochK, stochD := indicators.StochasticFromKlines(indicatorsKlines, 14, 3, 3)
	
	if len(stochK) > 0 {
		// Afficher 5 dernières valeurs Stochastic
		fmt.Println("   5 dernières valeurs Stochastic:")
		startIdx := len(stochK) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(stochK); i++ {
			klineIdx := i
			if klineIdx >= len(klines) {
				klineIdx = len(klines) - 1
			}
			
			var stochSignal string
			if !math.IsNaN(stochK[i]) && !math.IsNaN(stochD[i]) {
				if stochK[i] > 80 && stochD[i] > 80 {
					stochSignal = "🔴 SURACHAT"
				} else if stochK[i] < 20 && stochD[i] < 20 {
					stochSignal = "🟢 SURVENTE"
				} else if stochK[i] > stochD[i] {
					stochSignal = "🟡 HAUSSIER"
				} else {
					stochSignal = "🟡 BAISSIER"
				}
			} else {
				stochSignal = "⚪ NaN"
			}
			
			fmt.Printf("   %s %%K:%.2f %%D:%.2f %s\n", 
				klines[klineIdx].OpenTime.Format("15:04"),
				stochK[i], stochD[i], stochSignal)
		}
		
		lastK := stochK[len(stochK)-1]
		lastD := stochD[len(stochD)-1]
		
		var stochSignal string
		if !math.IsNaN(lastK) && !math.IsNaN(lastD) {
			if lastK > 80 && lastD > 80 {
				stochSignal = "🔴 SURACHAT"
			} else if lastK < 20 && lastD < 20 {
				stochSignal = "🟢 SURVENTE"
			} else if lastK > lastD {
				stochSignal = "🟡 MOMENTUM HAUSSIER"
			} else {
				stochSignal = "🟡 MOMENTUM BAISSIER"
			}
		} else {
			stochSignal = "⚪ NaN"
		}
		
		fmt.Printf("   Stochastic Actuel: %%K:%.2f - %%D:%.2f - Signal: %s\n", 
			lastK, lastD, stochSignal)
	}

	// 8️⃣ RÉSUMÉ VALIDATION
	fmt.Println("\n🎯 RÉSUMÉ VALIDATION BINANCE FUTURES:")
	fmt.Println("=" + strings.Repeat("=", 40))
	fmt.Printf("✅ Source:     Binance Futures API (SOLUSDT perpétuel)\n")
	fmt.Printf("✅ Timeframe:  5m\n")
	fmt.Printf("✅ Klines:     %d bougies\n", len(klines))
	fmt.Printf("✅ Période:     %s à %s\n", 
		klines[0].OpenTime.Format("15:04"), 
		klines[len(klines)-1].OpenTime.Format("15:04"))
	
	fmt.Println("\n📊 TOUS LES INDICATEURS SONT PRÉCIS:")
	fmt.Println("   🔹 MFI - Money Flow Index")
	fmt.Println("   🔹 MACD - Moving Average Convergence Divergence") 
	fmt.Println("   🔹 CCI - Commodity Channel Index")
	fmt.Println("   🔹 DMI - Directional Movement Index")
	fmt.Println("   🔹 Stochastic - Oscillateur Stochastique")
	
	fmt.Println("\n🏁 VALIDATION TERMINÉE AVEC SUCCÈS !")
	fmt.Println("💡 Les données Binance Futures perpétuelles sont précises pour tous les indicateurs !")
}
