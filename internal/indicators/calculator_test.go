package indicators

import (
	"testing"
	"time"
)

func TestCalculate_BasicFunctionality(t *testing.T) {
	// Test basic Calculate function
	klines := generateTestKlines(50)
	
	request := &CalculationRequest{
		Symbol:       "SOLUSDT",
		Timeframe:    "5m",
		CurrentTime:  time.Now().UnixMilli(),
		CandleWindow: klines,
		RequestID:    "test-basic",
	}
	
	response := Calculate(request)
	
	if !response.Success {
		t.Errorf("Expected success, got: %v", response.Error)
	}
	
	if response.Results == nil {
		t.Error("Expected results not nil")
	}
	
	if response.Results.MACD == nil {
		t.Error("Expected MACD results")
	}
	
	if response.Results.CCI == nil {
		t.Error("Expected CCI results")
	}
	
	if response.Results.DMI == nil {
		t.Error("Expected DMI results")
	}
}

func TestCalculate_EmptyKlines(t *testing.T) {
	request := &CalculationRequest{
		Symbol:       "SOLUSDT",
		CandleWindow: []Kline{},
		RequestID:    "test-empty",
	}
	
	response := Calculate(request)
	
	if response.Success {
		t.Error("Expected failure for empty klines")
	}
	
	if response.Error == nil {
		t.Error("Expected error for empty klines")
	}
}

func TestDetectMACDCrossover_CrossUp(t *testing.T) {
	macd := []float64{-0.5, 0.2}
	signal := []float64{-0.3, 0.1}
	
	result := detectMACDCrossover(macd, signal)
	
	if result != CrossUp {
		t.Errorf("Expected CrossUp, got %v", result)
	}
}

func TestDetectMACDCrossover_CrossDown(t *testing.T) {
	macd := []float64{0.5, -0.2}
	signal := []float64{0.3, -0.1}
	
	result := detectMACDCrossover(macd, signal)
	
	if result != CrossDown {
		t.Errorf("Expected CrossDown, got %v", result)
	}
}

func TestDetectMACDCrossover_NoCross(t *testing.T) {
	macd := []float64{0.5, 0.6}
	signal := []float64{0.3, 0.4}
	
	result := detectMACDCrossover(macd, signal)
	
	if result != NoCrossover {
		t.Errorf("Expected NoCrossover, got %v", result)
	}
}

func TestDetectMACDCrossover_InsufficientData(t *testing.T) {
	macd := []float64{0.5}
	signal := []float64{0.3}
	
	result := detectMACDCrossover(macd, signal)
	
	if result != NoCrossover {
		t.Errorf("Expected NoCrossover for insufficient data, got %v", result)
	}
}

func TestClassifyCCIZone_Oversold(t *testing.T) {
	result := classifyCCIZone(-150)
	
	if result != CCIOversold {
		t.Errorf("Expected CCIOversold, got %v", result)
	}
}

func TestClassifyCCIZone_Overbought(t *testing.T) {
	result := classifyCCIZone(150)
	
	if result != CCIOverbought {
		t.Errorf("Expected CCIOverbought, got %v", result)
	}
}

func TestClassifyCCIZone_Normal(t *testing.T) {
	result := classifyCCIZone(50)
	
	if result != CCINormal {
		t.Errorf("Expected CCINormal, got %v", result)
	}
}

func generateTestKlines(count int) []Kline {
	klines := make([]Kline, count)
	baseTime := time.Now().UnixMilli()
	basePrice := 100.0
	
	for i := 0; i < count; i++ {
		price := basePrice + float64(i%10-5)
		klines[i] = Kline{
			Timestamp: baseTime + int64(i*60000),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price + 0.1,
			Volume:    1000 + float64(i%100),
		}
	}
	
	return klines
}
