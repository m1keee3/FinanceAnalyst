package utils

import "math"

// ZNormalize performs Z-score normalization on the input slice:
// subtracts the mean and divides by the standard deviation.
// If the standard deviation is zero, returns a zeroed slice of the same length.
func ZNormalize(data []float64) []float64 {
	n := len(data)
	if n == 0 {
		return nil
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(n)

	var sqSum float64
	for _, v := range data {
		diff := v - mean
		sqSum += diff * diff
	}
	variance := sqSum / float64(n)
	stddev := math.Sqrt(variance)

	if stddev == 0 {
		out := make([]float64, n)
		return out
	}

	out := make([]float64, n)
	for i, v := range data {
		out[i] = (v - mean) / stddev
	}
	return out
}

// Resample linearly interpolates the input slice to the desired target length.
// If the input length is 0 or targetLen < 1, returns nil.
// If targetLen == 1, returns slice with the average of data.
func Resample(data []float64, targetLen int) []float64 {
	m := len(data)
	if m == 0 || targetLen < 1 {
		return nil
	}
	if targetLen == 1 {
		sum := 0.0
		for _, v := range data {
			sum += v
		}
		return []float64{sum / float64(m)}
	}

	out := make([]float64, targetLen)
	scale := float64(m-1) / float64(targetLen-1)
	for i := 0; i < targetLen; i++ {
		pos := scale * float64(i)
		idx := int(math.Floor(pos))
		if idx >= m-1 {
			out[i] = data[m-1]
			continue
		}
		frac := pos - float64(idx)
		out[i] = data[idx] + (data[idx+1]-data[idx])*frac
	}
	return out
}

// LbKeoghEnvelope создает огибающие (envelope) для нижней и верхней границ
func LbKeoghEnvelope(s []float64, resampleLen int) (lower, upper []float64) {
	n := len(s)
	lower = make([]float64, n)
	upper = make([]float64, n)
	window := int(math.Floor(float64(resampleLen) * (1 - 0.9)))
	for i := range s {
		l, u := s[i], s[i]
		for j := max(0, i-window); j < min(n, i+window); j++ {
			if s[j] < l {
				l = s[j]
			}
			if s[j] > u {
				u = s[j]
			}
		}
		lower[i] = l
		upper[i] = u
	}
	return
}

// LbKeoghDistance вычисляет LB_Keogh расстояние между candidate и target
func LbKeoghDistance(candidate, lower, upper, target []float64) float64 {
	var sum float64
	for i := range candidate {
		if target[i] > upper[i] {
			sum += (target[i] - upper[i]) * (target[i] - upper[i])
		} else if target[i] < lower[i] {
			sum += (lower[i] - target[i]) * (lower[i] - target[i])
		}
	}
	return math.Sqrt(sum)
}

// DTW вычисляет Dynamic Time Warping расстояние между двумя временными рядами
// с ранней остановкой, если стоимость превышает maxCost
func DTW(a, b []float64, maxCost float64) float64 {
	n, m := len(a), len(b)
	const inf = 1e9
	prev := make([]float64, m+1)
	cur := make([]float64, m+1)
	for j := range prev {
		prev[j] = inf
	}
	prev[0] = 0

	for i := 1; i <= n; i++ {
		cur[0] = inf
		rowMin := inf
		for j := 1; j <= m; j++ {
			cost := math.Abs(a[i-1] - b[j-1])
			minPrev := prev[j]
			if prev[j-1] < minPrev {
				minPrev = prev[j-1]
			}
			if cur[j-1] < minPrev {
				minPrev = cur[j-1]
			}
			cur[j] = cost + minPrev
			if cur[j] < rowMin {
				rowMin = cur[j]
			}
		}
		if rowMin > maxCost {
			return -1
		}
		prev, cur = cur, prev
	}
	return prev[m]
}
