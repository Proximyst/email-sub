package batching

func Batch[T any](items []T, batchSize int) [][]T {
	if batchSize <= 0 {
		panic("batch size must be greater than 0")
	}
	if len(items) == 0 {
		return nil
	}

	var batches [][]T
	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))
		batches = append(batches, items[i:end])
	}
	return batches
}
