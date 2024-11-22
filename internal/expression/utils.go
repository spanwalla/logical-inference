package expression

const invalidIdx = -1

func increaseIdx(idx, offset int) int {
	if idx == invalidIdx {
		return invalidIdx
	}
	return idx + offset
}

func decreaseIdx(idx, offset int) int {
	if idx == invalidIdx || offset > idx {
		return invalidIdx
	}
	return idx - offset
}
