package hashgrid

// cell holds x in the high 32 bits and y in the low.
type cell uint64

func packCell(x, y int32) cell {
	return cell(uint64(uint32(x))<<32 | uint64(uint32(y)))
}
