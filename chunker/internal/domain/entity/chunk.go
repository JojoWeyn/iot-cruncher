package entity

type Chunk struct {
	JobID         string
	ChunkID       int
	PayloadURL    string
	EncryptFields []string
}
