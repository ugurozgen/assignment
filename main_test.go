package main

import (
	"testing"
)

func TestCalculatePacksFunction(t *testing.T) {
	testCases := []struct {
		orderItems    int
		expectedPacks map[int]int
	}{
		{500, map[int]int{5000: 0, 2000: 0, 1000: 0, 500: 1, 250: 0}},
		{501, map[int]int{5000: 0, 2000: 0, 1000: 0, 500: 1, 250: 1}},
		{251, map[int]int{5000: 0, 2000: 0, 1000: 0, 500: 1, 250: 0}},
		{12001, map[int]int{5000: 2, 2000: 1, 1000: 0, 500: 0, 250: 1}},
		{1, map[int]int{5000: 0, 2000: 0, 1000: 0, 500: 0, 250: 1}},
		{250, map[int]int{5000: 0, 2000: 0, 1000: 0, 500: 0, 250: 1}},
	}

	for _, testCase := range testCases {
		result, _ := calculatePacks(testCase.orderItems)
		equal(result, testCase.expectedPacks)
	}
}

func equal(actual, exptected map[int]int) bool {
	for pack := range actual {
		if actual[pack] != exptected[pack] {
			return false
		}
	}
	return true
}
