package main

// sma-x

type SMA struct {
	len  int
	data []float64
	sum  float64
}

func NewSMA(len int) *SMA {
	sma := new(SMA)
	sma.len = len
	return sma
}

func (sma *SMA) Add(n float64) float64 {
	if len(sma.data) >= sma.len {
		sma.sum -= sma.data[0]
		sma.data = sma.data[1:]
	}

	sma.data = append(sma.data, n)
	sma.sum += n
	return sma.sum / float64(len(sma.data))
}
