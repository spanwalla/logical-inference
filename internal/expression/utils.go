package expression

const invalidIdx = ^uint(0)

func increaseIdx(idx, offset uint) uint {
	if idx == invalidIdx {
		return invalidIdx
	}
	return idx + offset
}

func decreaseIdx(idx, offset uint) uint {
	if idx == invalidIdx || offset > idx {
		return invalidIdx
	}
	return idx - offset
}
