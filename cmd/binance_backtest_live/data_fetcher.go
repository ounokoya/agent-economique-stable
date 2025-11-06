package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/engine"
)

// DataFetcher handles Binance API data retrieval for backtest simulation
type DataFetcher struct {
	client     *binance.Client
	symbol     string
	timeframe1 string // Primary timeframe (e.g., 5m)
	timeframe2 string // Secondary timeframe (e.g., 15m)
	verbose    bool
}

// HistoricalDataSet contains all fetched historical data
type HistoricalDataSet struct {
	Klines1        []binance.Kline    // Primary timeframe klines
	Klines2        []binance.Kline    // Secondary timeframe klines
	Trades         []engine.Trade     // Simulated trades from klines
	StartTime      time.Time          // Data start time
	EndTime        time.Time          // Data end time
	TotalCandles1  int                // Count of primary timeframe candles
	TotalCandles2  int                // Count of secondary timeframe candles
	TotalTrades    int                // Count of simulated trades
}

// NewDataFetcher creates a new data fetcher instance
func NewDataFetcher(client *binance.Client, symbol, tf1, tf2 string, verbose bool) *DataFetcher {
	return &DataFetcher{
		client:     client,
		symbol:     symbol,
		timeframe1: tf1,
		timeframe2: tf2,
		verbose:    verbose,
	}
}

// FetchHistoricalData retrieves all required historical data from Binance
func (df *DataFetcher) FetchHistoricalData(ctx context.Context, daysBack int) (*HistoricalDataSet, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -daysBack)
	
	fmt.Printf("📡 Fetching Binance data for %s (%d days)\n", df.symbol, daysBack)
	fmt.Printf("📅 Period: %s → %s\n", 
		startTime.Format("2006-01-02 15:04"), 
		endTime.Format("2006-01-02 15:04"))
	
	dataset := &HistoricalDataSet{
		StartTime: startTime,
		EndTime:   endTime,
	}
	
	// Step 1: Fetch primary timeframe klines (e.g., 5m)
	if err := df.fetchKlines(ctx, df.timeframe1, &dataset.Klines1, "PRIMARY"); err != nil {
		return nil, fmt.Errorf("failed to fetch primary klines (%s): %w", df.timeframe1, err)
	}
	dataset.TotalCandles1 = len(dataset.Klines1)
	
	// Step 2: Fetch secondary timeframe klines (e.g., 15m)
	if err := df.fetchKlines(ctx, df.timeframe2, &dataset.Klines2, "SECONDARY"); err != nil {
		return nil, fmt.Errorf("failed to fetch secondary klines (%s): %w", df.timeframe2, err)
	}
	dataset.TotalCandles2 = len(dataset.Klines2)
	
	// Step 3: Fetch real historical trades from Binance
	if err := df.fetchHistoricalTrades(ctx, df.timeframe1, &dataset.Trades); err != nil {
		return nil, fmt.Errorf("failed to fetch historical trades: %w", err)
	}
	dataset.TotalTrades = len(dataset.Trades)
	
	// Step 4: Validate and sort data chronologically
	if err := df.validateAndSortData(dataset); err != nil {
		return nil, fmt.Errorf("data validation failed: %w", err)
	}
	
	df.logDatasetSummary(dataset)
	return dataset, nil
}

// fetchKlines retrieves klines for a specific timeframe
func (df *DataFetcher) fetchKlines(ctx context.Context, timeframe string, klines *[]binance.Kline, label string) error {
	// Calculate limit based on timeframe to get ~10 days of data
	limit := df.calculateLimit(timeframe, 10)
	
	fmt.Printf("📊 Fetching %s klines (%s) - limit: %d\n", label, timeframe, limit)
	
	// Fetch from Binance API
	fetchedKlines, err := df.client.GetKlines(ctx, df.symbol, timeframe, limit)
	if err != nil {
		return fmt.Errorf("Binance API error: %w", err)
	}
	
	if len(fetchedKlines) == 0 {
		return fmt.Errorf("no klines returned for %s %s", df.symbol, timeframe)
	}
	
	*klines = fetchedKlines
	
	if df.verbose {
		first := fetchedKlines[0]
		last := fetchedKlines[len(fetchedKlines)-1]
		fmt.Printf("   ✅ %d klines fetched: %s → %s\n", 
			len(fetchedKlines),
			first.OpenTime.Format("02/01 15:04"),
			last.CloseTime.Format("02/01 15:04"))
	}
	
	return nil
}

// calculateLimit determines how many klines to fetch for given timeframe and days
func (df *DataFetcher) calculateLimit(timeframe string, days int) int {
	// Calculate klines per day for each timeframe
	klinesPerDay := map[string]int{
		"1m":  1440, // 24*60
		"5m":  288,  // 24*60/5
		"15m": 96,   // 24*60/15
		"1h":  24,   // 24
		"4h":  6,    // 24/4
		"1d":  1,    // 1
	}
	
	perDay, exists := klinesPerDay[timeframe]
	if !exists {
		perDay = 288 // Default to 5m
	}
	
	// Add 20% buffer for safety and weekends
	total := int(float64(perDay*days) * 1.2)
	
	// Binance API limits
	if total > 1500 {
		total = 1500
	}
	
	return total
}

// fetchHistoricalTrades retrieves real historical trades from Binance
func (df *DataFetcher) fetchHistoricalTrades(ctx context.Context, timeframe string, trades *[]engine.Trade) error {
	// Calculate limit for trades (approximately 8x more trades than klines)
	klinesLimit := df.calculateLimit(timeframe, 10)
	tradesLimit := klinesLimit * 8
	
	fmt.Printf("💰 Fetching real historical trades (%s) - limit: %d\n", timeframe, tradesLimit)
	
	// Fetch historical trades from Binance (improved from klines)
	binanceTrades, err := df.client.GetHistoricalAggTrades(ctx, df.symbol, timeframe, klinesLimit)
	if err != nil {
		return fmt.Errorf("Binance trades API error: %w", err)
	}
	
	if len(binanceTrades) == 0 {
		return fmt.Errorf("no trades returned for %s %s", df.symbol, timeframe)
	}
	
	// Convert Binance trades to engine trades
	engineTrades := make([]engine.Trade, len(binanceTrades))
	for i, trade := range binanceTrades {
		engineTrades[i] = engine.Trade{
			Timestamp:    trade.Time.UnixMilli(),
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			IsBuyerMaker: trade.IsBuyerMaker,
		}
	}
	
	*trades = engineTrades
	
	if df.verbose {
		first := binanceTrades[0]
		last := binanceTrades[len(binanceTrades)-1]
		fmt.Printf("   ✅ %d real trades fetched: %s → %s\n", 
			len(binanceTrades),
			first.Time.Format("02/01 15:04:05"),
			last.Time.Format("02/01 15:04:05"))
		fmt.Printf("   💲 Price range: %.2f → %.2f\n", first.Price, last.Price)
	}
	
	return nil
}

// validateAndSortData ensures data integrity and chronological order
func (df *DataFetcher) validateAndSortData(dataset *HistoricalDataSet) error {
	fmt.Printf("🔍 Validating and sorting data chronologically...\n")
	
	// Validate we have data
	if len(dataset.Klines1) == 0 {
		return fmt.Errorf("no primary klines data")
	}
	if len(dataset.Trades) == 0 {
		return fmt.Errorf("no trades data generated")
	}
	
	// Sort klines chronologically
	sort.Slice(dataset.Klines1, func(i, j int) bool {
		return dataset.Klines1[i].OpenTime.Before(dataset.Klines1[j].OpenTime)
	})
	
	if len(dataset.Klines2) > 0 {
		sort.Slice(dataset.Klines2, func(i, j int) bool {
			return dataset.Klines2[i].OpenTime.Before(dataset.Klines2[j].OpenTime)
		})
	}
	
	// Sort trades chronologically (CRITICAL for temporal engine)
	sort.Slice(dataset.Trades, func(i, j int) bool {
		return dataset.Trades[i].Timestamp < dataset.Trades[j].Timestamp
	})
	
	// Validate chronological order
	for i := 1; i < len(dataset.Trades); i++ {
		if dataset.Trades[i].Timestamp < dataset.Trades[i-1].Timestamp {
			return fmt.Errorf("trades not in chronological order at index %d", i)
		}
	}
	
	// Update actual time range based on data
	if len(dataset.Klines1) > 0 {
		dataset.StartTime = dataset.Klines1[0].OpenTime
		dataset.EndTime = dataset.Klines1[len(dataset.Klines1)-1].CloseTime
	}
	
	if df.verbose {
		fmt.Printf("   ✅ Data validated and sorted chronologically\n")
		fmt.Printf("   📊 Actual range: %s → %s\n", 
			dataset.StartTime.Format("02/01 15:04"), 
			dataset.EndTime.Format("02/01 15:04"))
	}
	
	return nil
}

// logDatasetSummary displays comprehensive data summary
func (df *DataFetcher) logDatasetSummary(dataset *HistoricalDataSet) {
	fmt.Printf("\n📋 DATA FETCH SUMMARY - %s\n", df.symbol)
	fmt.Printf("================================================================================\n")
	fmt.Printf("📅 Time Range:     %s → %s\n", 
		dataset.StartTime.Format("2006-01-02 15:04"), 
		dataset.EndTime.Format("2006-01-02 15:04"))
	fmt.Printf("⏱️  Duration:       %.1f days\n", 
		dataset.EndTime.Sub(dataset.StartTime).Hours()/24)
	fmt.Printf("📊 Klines (%s):    %d candles\n", df.timeframe1, dataset.TotalCandles1)
	fmt.Printf("📊 Klines (%s):   %d candles\n", df.timeframe2, dataset.TotalCandles2)
	fmt.Printf("💰 Simulated Trades: %d trades\n", dataset.TotalTrades)
	
	if len(dataset.Klines1) > 0 {
		first := dataset.Klines1[0]
		last := dataset.Klines1[len(dataset.Klines1)-1]
		priceChange := ((last.Close - first.Open) / first.Open) * 100
		fmt.Printf("💲 Price Movement:  %.2f → %.2f (%.2f%%)\n", 
			first.Open, last.Close, priceChange)
	}
	
	fmt.Printf("================================================================================\n\n")
}

// ConvertToEngineFormat converts Binance data to engine-compatible format with multi-timeframes
func (df *DataFetcher) ConvertToEngineFormat(dataset *HistoricalDataSet) (*engine.InitialData, error) {
	if len(dataset.Klines1) == 0 {
		return nil, fmt.Errorf("no primary klines available for conversion")
	}
	
	if len(dataset.Trades) == 0 {
		return nil, fmt.Errorf("no trades available for conversion")
	}
	
	// Convert multi-timeframes to engine format
	multiTFKlines := make(map[string][]engine.Kline)
	
	// Convert primary timeframe (tf1)
	engineKlines1 := make([]engine.Kline, len(dataset.Klines1))
	for i, k := range dataset.Klines1 {
		engineKlines1[i] = engine.Kline{
			Timestamp: k.OpenTime.UnixMilli(),
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}
	multiTFKlines[df.timeframe1] = engineKlines1
	
	// Convert secondary timeframe (tf2) if available
	if len(dataset.Klines2) > 0 {
		engineKlines2 := make([]engine.Kline, len(dataset.Klines2))
		for i, k := range dataset.Klines2 {
			engineKlines2[i] = engine.Kline{
				Timestamp: k.OpenTime.UnixMilli(),
				Open:      k.Open,
				High:      k.High,
				Low:       k.Low,
				Close:     k.Close,
				Volume:    k.Volume,
			}
		}
		multiTFKlines[df.timeframe2] = engineKlines2
	}
	
	if df.verbose {
		fmt.Printf("🔄 Converting to multi-timeframe engine format\n")
		fmt.Printf("   📊 DEBUG: tf1='%s', tf2='%s'\n", df.timeframe1, df.timeframe2)
		fmt.Printf("   📊 DEBUG: Klines1 count: %d, Klines2 count: %d\n", len(dataset.Klines1), len(dataset.Klines2))
		fmt.Printf("   📊 DEBUG: MultiTF map size: %d\n", len(multiTFKlines))
		for tf, klines := range multiTFKlines {
			fmt.Printf("   📊 %s: %d klines (%s → %s)\n", 
				tf, len(klines),
				time.UnixMilli(klines[0].Timestamp).Format("02/01 15:04"),
				time.UnixMilli(klines[len(klines)-1].Timestamp).Format("02/01 15:04"))
		}
		fmt.Printf("   💰 %d trades total\n", len(dataset.Trades))
	}
	
	return &engine.InitialData{
		Trades:              dataset.Trades,
		RecentKlinesMultiTF: multiTFKlines,
		RecentTrades:        dataset.Trades[max(0, len(dataset.Trades)-100):], // Last 100 trades
		
		// Legacy compatibility: use primary timeframe as single-TF fallback
		RecentKlines:        engineKlines1,
	}, nil
}

// GetTimeframeDuration returns duration for a timeframe string
func GetTimeframeDuration(timeframe string) time.Duration {
	switch timeframe {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 5 * time.Minute // Default to 5m
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
