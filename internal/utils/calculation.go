package utils

import (
	"fmt"
	"strconv"
)

func ConvertHourlyToMonthly(hourlyCostStr string) string {
	hourlyCost, err := strconv.ParseFloat(hourlyCostStr, 64)
	if err != nil {
		panic(err)
	}

	// 730 hours in a month
	return fmt.Sprintf("%.2f", hourlyCost*730)
}
