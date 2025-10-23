package entity

import "time"

type ChunkResult struct {
	JobID     string
	ChunkID   int
	Stats     ChunkStats
	Anomalies []SensorReading
}

type ChunkStats struct {
	SensorID    string
	Temperature Stats
	Humidity    Stats
	Pressure    Stats
}

type Stats struct {
	Min  float64
	Max  float64
	Mean float64
	Std  float64
}

type SensorReading struct {
	Timestamp   time.Time
	SensorID    string
	Temperature float64
	Humidity    float64
	Pressure    float64
}
