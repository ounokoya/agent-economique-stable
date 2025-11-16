package ban_fin_momentium

import (
	"fmt"
	"math"

	"agent-economique/internal/signals"
)

type WindowOpenMode string

const (
	WindowOpenModeAuto       WindowOpenMode = "auto"
	WindowOpenModeCross      WindowOpenMode = "cross"
	WindowOpenModeBar        WindowOpenMode = "bar"
	WindowOpenModeStochCross WindowOpenMode = "stoch_cross"
	WindowOpenModeVWMACross  WindowOpenMode = "vwma_cross"
	WindowOpenModeMACDCross  WindowOpenMode = "macd_cross"
)

type WindowConfig struct {
	WindowBars int
	OpenMode   WindowOpenMode
}

type windowState struct {
	active   bool
	dir      signals.SignalType
	startIdx int
	deadline int

	lastStochCross int // +1 up, -1 down, 0 none
	lastVWMACross  int
	lastMACDCross  int
}

// WindowFinder applique la logique de fen
// tre optionnelle par-dessus un Generator BAN_FIN_MOMENTIUM.
// Il ne modifie pas le Generator lui-mame et n'est utilise que si EnableWindowMode est active cf4te appelant.
type WindowFinder struct {
	gen    *Generator
	cfg    WindowConfig
	longW  windowState
	shortW windowState
}

func NewWindowFinder(gen *Generator, cfg WindowConfig) *WindowFinder {
	bars := cfg.WindowBars
	if bars <= 0 {
		bars = 10
	}
	mode := cfg.OpenMode
	if mode != WindowOpenModeAuto && mode != WindowOpenModeCross && mode != WindowOpenModeBar &&
		mode != WindowOpenModeStochCross && mode != WindowOpenModeVWMACross && mode != WindowOpenModeMACDCross {
		mode = WindowOpenModeAuto
	}
	return &WindowFinder{
		gen: gen,
		cfg: WindowConfig{WindowBars: bars, OpenMode: mode},
	}
}

// Step evalue la bougie idx (derniare bougie ferme) dans une logique de fenetre.
// klines est le flux complet de bougies (fermees et en cours), idx est l'index de la bougie fermee ne analyser.
// Retourne au plus un signal directionnel ENTRY.
func (f *WindowFinder) Step(klines []signals.Kline, idx int) (*signals.Signal, error) {
	g := f.gen
	if idx <= 0 || idx >= len(klines) {
		return nil, fmt.Errorf("invalid idx %d for window step", idx)
	}

	// Validation des indicateurs comme dans EvaluateLast
	if idx >= len(g.atr) || math.IsNaN(g.atr[idx]) {
		return nil, nil
	}
	if idx >= len(g.volSMA) || math.IsNaN(g.volSMA[idx]) {
		return nil, nil
	}
	stochExtGatesEnabled := g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
		g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate
	needStoch := g.cfg.EnableStochCross || stochExtGatesEnabled
	needMFI := g.cfg.EnableMFIGate
	needCCI := g.cfg.EnableCCIGate
	needVWMA := g.cfg.EnableVWMATrendGate || g.cfg.EnableVWMACross
	needMACD := g.cfg.EnableMACDCrossGate || g.cfg.EnableMACDSignGate || g.cfg.EnableMACDHistGate
	// Si aucun gate n'est actif (bar, Stoch, MFI, CCI, VWMA, MACD), ne pas produire de signaux en mode fenêtre.
	hasAnyGate := g.cfg.EnableBarGate || g.cfg.EnableStochCross || stochExtGatesEnabled || needMFI || needCCI || needVWMA || needMACD
	if !hasAnyGate {
		return nil, nil
	}
	if needStoch {
		if idx >= len(g.stochK) || math.IsNaN(g.stochK[idx]) {
			return nil, nil
		}
		if g.cfg.EnableStochCross {
			if idx >= len(g.stochD) || math.IsNaN(g.stochD[idx]) {
				return nil, nil
			}
		}
	}
	if needMFI {
		if idx >= len(g.mfi) || math.IsNaN(g.mfi[idx]) {
			return nil, nil
		}
	}
	if needCCI {
		if idx >= len(g.cci) || math.IsNaN(g.cci[idx]) {
			return nil, nil
		}
	}
	if needVWMA {
		if idx >= len(g.vwmaFast) || math.IsNaN(g.vwmaFast[idx]) {
			return nil, nil
		}
		if idx >= len(g.vwmaSlow) || math.IsNaN(g.vwmaSlow[idx]) {
			return nil, nil
		}
	}
	if needMACD {
		if idx >= len(g.macdLine) || math.IsNaN(g.macdLine[idx]) {
			return nil, nil
		}
		if idx >= len(g.macdSignal) || math.IsNaN(g.macdSignal[idx]) {
			return nil, nil
		}
		if idx >= len(g.macdHist) || math.IsNaN(g.macdHist[idx]) {
			return nil, nil
		}
	}

	body, isGreen, isRed, volNorm := g.bodyVolumeColorAt(klines, idx)
	b2atr := body / g.atr[idx]
	v2sma := volNorm / g.volSMA[idx]

	// Croisements locaux sur cette bougie
	stochUp, stochDown := false, false
	if g.cfg.EnableStochCross {
		up, down := g.stochCrossAt(idx)
		stochUp, stochDown = up, down
	}
	vwUp, vwDown := false, false
	if g.cfg.EnableVWMACross {
		up, down := g.vwmaCrossAt(idx)
		vwUp, vwDown = up, down
	}
	macdUp, macdDown := false, false
	if g.cfg.EnableMACDCrossGate {
		up, down := g.macdCrossAt(idx)
		macdUp, macdDown = up, down
	}

	// Trend via VWMA (CCI_trend est ge9re par la de9mo comme avant)
	trendOKLong := true
	trendOKShort := true
	if g.cfg.EnableVWMATrendGate {
		trendOKLong = g.vwmaFast[idx] > g.vwmaSlow[idx]
		trendOKShort = g.vwmaFast[idx] < g.vwmaSlow[idx]
	}

	// Analyse de barre
	barOKLong, barOKShort := true, true
	if g.cfg.EnableBarGate {
		barOKLong = isGreen && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
		barOKShort = isRed && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
	}

	// CCI court
	cciOKLong, cciOKShort := true, true
	if g.cfg.EnableCCIGate {
		cciVal := g.cci[idx]
		if g.cfg.EnableCCIOversoldLongGate {
			cciOKLong = cciOKLong && cciVal >= g.cfg.CCIOversoldLong
		}
		if g.cfg.EnableCCIOverboughtLongGate {
			cciOKLong = cciOKLong && cciVal <= g.cfg.CCIOverboughtLong
		}
		if g.cfg.EnableCCIOversoldShortGate {
			cciOKShort = cciOKShort && cciVal >= g.cfg.CCIOversoldShort
		}
		if g.cfg.EnableCCIOverboughtShortGate {
			cciOKShort = cciOKShort && cciVal <= g.cfg.CCIOverboughtShort
		}
	}

	// MFI
	mfiOKLong, mfiOKShort := true, true
	if g.cfg.EnableMFIGate {
		mfiVal := g.mfi[idx]
		if g.cfg.EnableMFIOversoldLongGate {
			mfiOKLong = mfiOKLong && mfiVal >= g.cfg.MFIOversoldLong
		}
		if g.cfg.EnableMFIOverboughtLongGate {
			mfiOKLong = mfiOKLong && mfiVal <= g.cfg.MFIOverboughtLong
		}
		if g.cfg.EnableMFIOversoldShortGate {
			mfiOKShort = mfiOKShort && mfiVal >= g.cfg.MFIOversoldShort
		}
		if g.cfg.EnableMFIOverboughtShortGate {
			mfiOKShort = mfiOKShort && mfiVal <= g.cfg.MFIOverboughtShort
		}
	}

	// Stoch K extr0ames
	stochKOKLong, stochKOKShort := true, true
	if g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
		g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate {
		kVal := g.stochK[idx]
		if g.cfg.EnableStochKOversoldLongGate {
			stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOversoldLong
		}
		if g.cfg.EnableStochKOverboughtLongGate {
			stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOverboughtLong
		}
		if g.cfg.EnableStochKOversoldShortGate {
			stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOversoldShort
		}
		if g.cfg.EnableStochKOverboughtShortGate {
			stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOverboughtShort
		}
	}

	// MACD sign/histo
	macdSignOKLong, macdSignOKShort := true, true
	if g.cfg.EnableMACDSignGate {
		m := g.macdLine[idx]
		s := g.macdSignal[idx]
		macdSignOKLong = m < 0 && s < 0
		macdSignOKShort = m > 0 && s > 0
	}
	macdHistOKLong, macdHistOKShort := true, true
	if g.cfg.EnableMACDHistGate {
		h := g.macdHist[idx]
		macdHistOKLong = h > 0
		macdHistOKShort = h < 0
	}

	// Mise e jour / ouverture des fenetres
	crossGateEnabled := g.cfg.EnableStochCross || g.cfg.EnableVWMACross || g.cfg.EnableMACDCrossGate

	// Helpers croisement par direction
	longCrossEvent := (g.cfg.EnableStochCross && stochUp) || (g.cfg.EnableVWMACross && vwUp) || (g.cfg.EnableMACDCrossGate && macdUp)
	shortCrossEvent := (g.cfg.EnableStochCross && stochDown) || (g.cfg.EnableVWMACross && vwDown) || (g.cfg.EnableMACDCrossGate && macdDown)

	// Ouverture fenetre LONG
	if g.cfg.EnableOpenLong && !f.longW.active {
		shouldOpen := false
		switch f.cfg.OpenMode {
		case WindowOpenModeAuto:
			if crossGateEnabled {
				shouldOpen = longCrossEvent
			} else {
				shouldOpen = barOKLong
			}
		case WindowOpenModeCross:
			if crossGateEnabled {
				shouldOpen = longCrossEvent
			}
		case WindowOpenModeBar:
			shouldOpen = barOKLong
		case WindowOpenModeStochCross:
			if g.cfg.EnableStochCross {
				shouldOpen = stochUp
			}
		case WindowOpenModeVWMACross:
			if g.cfg.EnableVWMACross {
				shouldOpen = vwUp
			}
		case WindowOpenModeMACDCross:
			if g.cfg.EnableMACDCrossGate {
				shouldOpen = macdUp
			}
		}
		if shouldOpen {
			f.longW.active = true
			f.longW.dir = signals.SignalTypeLong
			f.longW.startIdx = idx
			f.longW.deadline = idx + f.cfg.WindowBars - 1
			// Initialiser les derniers crosses
			if g.cfg.EnableStochCross {
				if stochUp {
					f.longW.lastStochCross = 1
				} else if stochDown {
					f.longW.lastStochCross = -1
				}
			}
			if g.cfg.EnableVWMACross {
				if vwUp {
					f.longW.lastVWMACross = 1
				} else if vwDown {
					f.longW.lastVWMACross = -1
				}
			}
			if g.cfg.EnableMACDCrossGate {
				if macdUp {
					f.longW.lastMACDCross = 1
				} else if macdDown {
					f.longW.lastMACDCross = -1
				}
			}
		}
	}

	// Ouverture fenetre SHORT
	if g.cfg.EnableOpenShort && !f.shortW.active {
		shouldOpen := false
		switch f.cfg.OpenMode {
		case WindowOpenModeAuto:
			if crossGateEnabled {
				shouldOpen = shortCrossEvent
			} else {
				shouldOpen = barOKShort
			}
		case WindowOpenModeCross:
			if crossGateEnabled {
				shouldOpen = shortCrossEvent
			}
		case WindowOpenModeBar:
			shouldOpen = barOKShort
		case WindowOpenModeStochCross:
			if g.cfg.EnableStochCross {
				shouldOpen = stochDown
			}
		case WindowOpenModeVWMACross:
			if g.cfg.EnableVWMACross {
				shouldOpen = vwDown
			}
		case WindowOpenModeMACDCross:
			if g.cfg.EnableMACDCrossGate {
				shouldOpen = macdDown
			}
		}
		if shouldOpen {
			f.shortW.active = true
			f.shortW.dir = signals.SignalTypeShort
			f.shortW.startIdx = idx
			f.shortW.deadline = idx + f.cfg.WindowBars - 1
			if g.cfg.EnableStochCross {
				if stochUp {
					f.shortW.lastStochCross = 1
				} else if stochDown {
					f.shortW.lastStochCross = -1
				}
			}
			if g.cfg.EnableVWMACross {
				if vwUp {
					f.shortW.lastVWMACross = 1
				} else if vwDown {
					f.shortW.lastVWMACross = -1
				}
			}
			if g.cfg.EnableMACDCrossGate {
				if macdUp {
					f.shortW.lastMACDCross = 1
				} else if macdDown {
					f.shortW.lastMACDCross = -1
				}
			}
		}
	}

	// Expiration des fenetres
	if f.longW.active && idx > f.longW.deadline {
		f.longW = windowState{}
	}
	if f.shortW.active && idx > f.shortW.deadline {
		f.shortW = windowState{}
	}

	// Mise e jour des derniers crosses dans les fenetres actives
	if f.longW.active {
		if g.cfg.EnableStochCross {
			if stochUp {
				f.longW.lastStochCross = 1
			} else if stochDown {
				f.longW.lastStochCross = -1
			}
		}
		if g.cfg.EnableVWMACross {
			if vwUp {
				f.longW.lastVWMACross = 1
			} else if vwDown {
				f.longW.lastVWMACross = -1
			}
		}
		if g.cfg.EnableMACDCrossGate {
			if macdUp {
				f.longW.lastMACDCross = 1
			} else if macdDown {
				f.longW.lastMACDCross = -1
			}
		}
	}
	if f.shortW.active {
		if g.cfg.EnableStochCross {
			if stochUp {
				f.shortW.lastStochCross = 1
			} else if stochDown {
				f.shortW.lastStochCross = -1
			}
		}
		if g.cfg.EnableVWMACross {
			if vwUp {
				f.shortW.lastVWMACross = 1
			} else if vwDown {
				f.shortW.lastVWMACross = -1
			}
		}
		if g.cfg.EnableMACDCrossGate {
			if macdUp {
				f.shortW.lastMACDCross = 1
			} else if macdDown {
				f.shortW.lastMACDCross = -1
			}
		}
	}

	// Validation des fenetres
	// On respecte la priorite LONG puis SHORT comme dans EvaluateLast (OPEN LONG avant OPEN SHORT).

	// LONG
	if f.longW.active && g.cfg.EnableOpenLong {
		crossOK := true
		if g.cfg.EnableStochCross {
			crossOK = crossOK && f.longW.lastStochCross == 1
		}
		if g.cfg.EnableVWMACross {
			crossOK = crossOK && f.longW.lastVWMACross == 1
		}
		if g.cfg.EnableMACDCrossGate {
			crossOK = crossOK && f.longW.lastMACDCross == 1
		}
		if crossOK && trendOKLong && stochKOKLong && barOKLong && cciOKLong && mfiOKLong && macdSignOKLong && macdHistOKLong {
			// Construire signal LONG
			sig := g.build(signals.SignalActionEntry, signals.SignalTypeLong, klines, idx, b2atr, v2sma, isGreen, isRed)
			// Une fois un signal valide, on reinitialise les fenetres
			f.longW = windowState{}
			f.shortW = windowState{}
			return sig, nil
		}
	}

	// SHORT
	if f.shortW.active && g.cfg.EnableOpenShort {
		crossOK := true
		if g.cfg.EnableStochCross {
			crossOK = crossOK && f.shortW.lastStochCross == -1
		}
		if g.cfg.EnableVWMACross {
			crossOK = crossOK && f.shortW.lastVWMACross == -1
		}
		if g.cfg.EnableMACDCrossGate {
			crossOK = crossOK && f.shortW.lastMACDCross == -1
		}
		if crossOK && trendOKShort && stochKOKShort && barOKShort && cciOKShort && mfiOKShort && macdSignOKShort && macdHistOKShort {
			sig := g.build(signals.SignalActionEntry, signals.SignalTypeShort, klines, idx, b2atr, v2sma, isGreen, isRed)
			f.longW = windowState{}
			f.shortW = windowState{}
			return sig, nil
		}
	}

	return nil, nil
}
