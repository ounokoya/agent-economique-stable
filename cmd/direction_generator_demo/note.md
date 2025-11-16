pour timeframe 1m vwma supérieur ex: 5, 10 a de meuilleur resultat
mais sur timeframe plus elevé les les valeur plus petite sont meilleur



	SYMBOL     = "SOL_USDT"
	TIMEFRAME  = "1m"
	NB_CANDLES = 500

	VWMA_RAPIDE           = 6 //3
	PERIODE_PENTE         = 4 //5
	SEUIL_PENTE_VWMA      = 0.1
	K_CONFIRMATION        = 3 // 2
	USE_DYNAMIC_THRESHOLD = true
	ATR_PERIODE           = 4    //4
	ATR_COEFFICIENT       = 0.25 //0.8