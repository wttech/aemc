package intsx

func MaxOf(args ...int) int {
	result := args[0]
	for _, arg := range args {
		if arg > result {
			result = arg
		}
	}
	return result
}
