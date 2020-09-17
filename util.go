package main

func minmax(array []int) (int, int) {
	if len(array) == 0 {
		return 0, 0
	}
	var max int = array[0]
	var min int = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}
