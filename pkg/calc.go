package pkg

import (
	"fmt"
	"math"
)

func Calc_diff(data []int) []int {
	// 累積だと見づらいので新規数に変換
	var npatients_new []int
	for i := 0; i < len(data)-1; i++ {
		npatients_new = append(npatients_new, data[i]-data[i+1])
	}
	return npatients_new
}

func Calc_sum(data []int) (int, error) {
	x := 0
	for _, k := range data {
		x = x + k
	}
	return x, nil
}

func Calc_average(data []int) (int, error) {
	x := 0
	for _, k := range data {
		x = x + k
	}
	if len(data) == 0 {
		err := fmt.Errorf("zero divided")
		return 0, err
	}
	x = x / len(data)
	return x, nil
}

func Calc_geometric_mean(data []float64) float64 {
	x := 1.0
	for _, k := range data {
		x = x * k
	}
	x = math.Pow(x, 1.0/float64(len(data)))
	return x
}

func Calc_week_average(npatients_new []int) ([]int, error) {
	var average_npatients_new []int
	days_week := 7
	for i := 0; i < len(npatients_new)-days_week+1; i++ {
		input := npatients_new[i : days_week+i]
		v, err := Calc_average(input)
		if err != nil {
			return nil, err
		}
		average_npatients_new = append(average_npatients_new, v)
	}
	return average_npatients_new, nil
}

func Calc_ratio(average_npatients_new []int) []float64 {
	var ratio_npatients []float64
	for i := 0; i < len(average_npatients_new)-1; i++ {
		v := float64(average_npatients_new[i]) / float64(average_npatients_new[i+1])
		ratio_npatients = append(ratio_npatients, v)
	}
	return ratio_npatients
}
