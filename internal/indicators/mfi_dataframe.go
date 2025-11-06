package indicators

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// MFIDataFrame calculates Money Flow Index using Go dataframes (pandas-like approach)
func MFIDataFrame(klines []Kline, period int) ([]float64, error) {
	if len(klines) < period {
		return nil, fmt.Errorf("insufficient data: need at least %d klines, got %d", period, len(klines))
	}

	// Convert klines to dataframe
	df := klinestoDataFrame(klines)

	// Calculate typical price (HLC3) manually
	highs := df.Col("High").Float()
	lows := df.Col("Low").Float()
	closes := df.Col("Close").Float()
	volumes := df.Col("Volume").Float()
	
	typicalPrices := make([]float64, len(klines))
	rawMoneyFlow := make([]float64, len(klines))
	
	for i := 0; i < len(klines); i++ {
		typicalPrices[i] = (highs[i] + lows[i] + closes[i]) / 3.0
		rawMoneyFlow[i] = typicalPrices[i] * volumes[i]
	}

	// Calculate money flow direction (positive/negative)
	positiveFlow := make([]float64, len(klines))
	negativeFlow := make([]float64, len(klines))

	for i := 1; i < len(typicalPrices); i++ {
		if typicalPrices[i] > typicalPrices[i-1] {
			positiveFlow[i] = rawMoneyFlow[i]
			negativeFlow[i] = 0
		} else if typicalPrices[i] < typicalPrices[i-1] {
			positiveFlow[i] = 0
			negativeFlow[i] = rawMoneyFlow[i]
		} else {
			positiveFlow[i] = 0
			negativeFlow[i] = 0
		}
	}

	// Calculate rolling sums for MFI
	mfiValues := make([]float64, len(klines))

	for i := period - 1; i < len(klines); i++ {
		// Calculate sums over the period
		var positiveSum, negativeSum float64
		
		for j := i - period + 1; j <= i; j++ {
			positiveSum += positiveFlow[j]
			negativeSum += negativeFlow[j]
		}

		// Calculate MFI
		if negativeSum == 0 {
			mfiValues[i] = 100.0
		} else {
			moneyFlowRatio := positiveSum / negativeSum
			mfiValues[i] = 100.0 - (100.0 / (1.0 + moneyFlowRatio))
		}
	}

	return mfiValues, nil
}

// MFIDataFrameSmoothed calculates smoothed MFI using dataframes with exponential smoothing
func MFIDataFrameSmoothed(klines []Kline, period int, smoothing float64) ([]float64, error) {
	// First calculate regular MFI
	regularMFI, err := MFIDataFrame(klines, period)
	if err != nil {
		return nil, err
	}

	// Apply exponential smoothing
	smoothedMFI := make([]float64, len(regularMFI))
	smoothedMFI[0] = regularMFI[0]

	alpha := 2.0 / (smoothing + 1.0)

	for i := 1; i < len(regularMFI); i++ {
		if regularMFI[i] != 0 {
			smoothedMFI[i] = alpha*regularMFI[i] + (1.0-alpha)*smoothedMFI[i-1]
		} else {
			smoothedMFI[i] = smoothedMFI[i-1]
		}
	}

	return smoothedMFI, nil
}

// MFIDataFrameEnhanced calculates enhanced MFI with multiple timeframe analysis
func MFIDataFrameEnhanced(klines []Kline, period int) (*MFIAnalysis, error) {
	df := klinestoDataFrame(klines)

	// Calculate multiple MFI variants
	standardMFI, err := MFIDataFrame(klines, period)
	if err != nil {
		return nil, err
	}

	smoothedMFI, err := MFIDataFrameSmoothed(klines, period, 5.0)
	if err != nil {
		return nil, err
	}

	// Calculate volume-weighted MFI
	volumeWeightedMFI, err := calculateVolumeWeightedMFI(df, period)
	if err != nil {
		return nil, err
	}

	return &MFIAnalysis{
		Standard:       standardMFI,
		Smoothed:       smoothedMFI,
		VolumeWeighted: volumeWeightedMFI,
		Period:         period,
	}, nil
}

// MFIAnalysis holds comprehensive MFI analysis results
type MFIAnalysis struct {
	Standard       []float64
	Smoothed       []float64
	VolumeWeighted []float64
	Period         int
}

// klinestoDataFrame converts klines slice to gota dataframe
func klinestoDataFrame(klines []Kline) dataframe.DataFrame {
	timestamps := make([]int64, len(klines))
	opens := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	closes := make([]float64, len(klines))
	volumes := make([]float64, len(klines))

	for i, k := range klines {
		timestamps[i] = k.Timestamp
		opens[i] = k.Open
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
		volumes[i] = k.Volume
	}

	df := dataframe.New(
		series.New(timestamps, series.Int, "Timestamp"),
		series.New(opens, series.Float, "Open"),
		series.New(highs, series.Float, "High"),
		series.New(lows, series.Float, "Low"),
		series.New(closes, series.Float, "Close"),
		series.New(volumes, series.Float, "Volume"),
	)

	return df
}

// calculateVolumeWeightedMFI calculates volume-weighted MFI for better accuracy
func calculateVolumeWeightedMFI(df dataframe.DataFrame, period int) ([]float64, error) {
	nrows, _ := df.Dims() // Use Dims() instead of Nrows()
	n := nrows
	mfiValues := make([]float64, n)

	closes := df.Col("Close").Float()
	highs := df.Col("High").Float()
	lows := df.Col("Low").Float()
	volumes := df.Col("Volume").Float()

	for i := period; i < n; i++ {
		var positiveFlow, negativeFlow float64
		var totalVolume float64

		for j := i - period + 1; j <= i; j++ {
			typicalPrice := (highs[j] + lows[j] + closes[j]) / 3.0
			prevTypicalPrice := (highs[j-1] + lows[j-1] + closes[j-1]) / 3.0

			rawMoneyFlow := typicalPrice * volumes[j]
			totalVolume += volumes[j]

			if typicalPrice > prevTypicalPrice {
				positiveFlow += rawMoneyFlow
			} else if typicalPrice < prevTypicalPrice {
				negativeFlow += rawMoneyFlow
			}
		}

		// Apply volume weighting
		if totalVolume > 0 {
			positiveFlow = positiveFlow / totalVolume * 100
			negativeFlow = negativeFlow / totalVolume * 100
		}

		if negativeFlow == 0 {
			mfiValues[i] = 100.0
		} else {
			moneyFlowRatio := positiveFlow / negativeFlow
			mfiValues[i] = 100.0 - (100.0 / (1.0 + moneyFlowRatio))
		}
	}

	return mfiValues, nil
}

// GetMFISignal returns trading signal based on MFI analysis
func (mfi *MFIAnalysis) GetMFISignal(index int) string {
	if index >= len(mfi.Standard) || index < 0 {
		return "NO_DATA"
	}

	standard := mfi.Standard[index]
	smoothed := mfi.Smoothed[index]

	// MFI overbought/oversold levels
	if standard > 80 && smoothed > 75 {
		return "STRONG_SELL"
	} else if standard > 70 {
		return "SELL"
	} else if standard < 20 && smoothed < 25 {
		return "STRONG_BUY"
	} else if standard < 30 {
		return "BUY"
	}

	return "NEUTRAL"
}

// PrintMFIAnalysis prints detailed MFI analysis
func (mfi *MFIAnalysis) PrintMFIAnalysis(exchange string, symbol string) {
	if len(mfi.Standard) == 0 {
		fmt.Printf("⚠️ No MFI data available for %s\n", exchange)
		return
	}

	lastIndex := len(mfi.Standard) - 1
	fmt.Printf("\n📊 MFI DATAFRAME ANALYSIS - %s (%s)\n", exchange, symbol)
	fmt.Printf("=====================================\n")
	fmt.Printf("Standard MFI:        %.2f\n", mfi.Standard[lastIndex])
	fmt.Printf("Smoothed MFI:        %.2f\n", mfi.Smoothed[lastIndex])
	fmt.Printf("Volume Weighted MFI: %.2f\n", mfi.VolumeWeighted[lastIndex])
	fmt.Printf("Signal:              %s\n", mfi.GetMFISignal(lastIndex))
	fmt.Printf("Period:              %d\n", mfi.Period)
}
