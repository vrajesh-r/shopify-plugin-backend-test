package dbhandlers

func ConvertDollarFloatToCents(dollars float64) int {
	cents := dollars * 100
	return int(cents)
}

func ConvertDollarFloatToCentsInt(dollars float64) int {
	cents := dollars * 100
	return int(cents)
}
