package atmcs

import (
	"github.com/dragonzurfer/revclose"
	"github.com/dragonzurfer/trader/executor"
)

type CandleAdapter struct {
	candles []executor.CandleLike
}

type RevCandleAdapter struct {
	candle executor.CandleLike
}

func (rca *RevCandleAdapter) GetHigh() float64 {
	return rca.candle.GetHigh()
}

func (rca *RevCandleAdapter) GetLow() float64 {
	return rca.candle.GetLow()
}

func (rca *RevCandleAdapter) GetClose() float64 {
	return rca.candle.GetClose()
}

func (rca *RevCandleAdapter) GetOpen() float64 {
	return rca.candle.GetOpen()
}

func (rca *RevCandleAdapter) GetOHLC() (float64, float64, float64, float64) {
	return rca.candle.GetOpen(), rca.candle.GetHigh(), rca.candle.GetLow(), rca.candle.GetClose()
}

func (ca *CandleAdapter) GetCandle(index int) revclose.RevCandle {
	return &RevCandleAdapter{candle: ca.candles[index]}
}

func (ca *CandleAdapter) GetCandlesLength() int {
	return len(ca.candles)
}

func NewCandleAdapter(candles []executor.CandleLike) *CandleAdapter {
	return &CandleAdapter{
		candles: candles,
	}
}
