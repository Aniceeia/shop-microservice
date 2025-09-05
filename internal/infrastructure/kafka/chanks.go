package kafka

const (
	MaxKafkaMessageSize = 50 * 1024 * 1024 // 50MB
	ChunkSize           = 1 * 1024 * 1024  // 1MB для чанков
)

type ChunkedMessage struct {
	ChunkIndex int    `json:"chunkIndex"`
	ChunkCount int    `json:"chunkCount"`
	Key        string `json:"key"`
	Value      []byte `json:"value"`
}

// MergeChunks объединяет части сообщения в одно целое
func MergeChunks(chunks [][]byte) []byte {
	if len(chunks) == 0 {
		return nil
	}

	var totalSize int
	for _, chunk := range chunks {
		totalSize += len(chunk)
	}

	result := make([]byte, 0, totalSize)
	for _, chunk := range chunks {
		result = append(result, chunk...)
	}
	return result
}

// AllChunksReceived проверяет, что все части сообщения пришли
func AllChunksReceived(chunks [][]byte) bool {
	if len(chunks) == 0 {
		return false
	}

	for _, chunk := range chunks {
		if chunk == nil {
			return false
		}
	}
	return true
}

// SplitIntoChunks разбивает сообщение на чанки заданного размера
func SplitIntoChunks(key string, value []byte, chunkSize int) ([]ChunkedMessage, error) {
	if len(value) == 0 {
		// Для пустых данных возвращаем один пустой чанк
		return []ChunkedMessage{
			{
				ChunkIndex: 0,
				ChunkCount: 1,
				Key:        key,
				Value:      []byte{},
			},
		}, nil
	}

	chunkCount := (len(value) + chunkSize - 1) / chunkSize
	chunks := make([]ChunkedMessage, chunkCount)

	for i := 0; i < chunkCount; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(value) {
			end = len(value)
		}

		chunks[i] = ChunkedMessage{
			ChunkIndex: i,
			ChunkCount: chunkCount,
			Key:        key,
			Value:      value[start:end],
		}
	}
	return chunks, nil
}

// ShouldChunk проверяет, нужно ли чанковать данные
func ShouldChunk(data []byte) bool {
	return len(data) > ChunkSize
}
