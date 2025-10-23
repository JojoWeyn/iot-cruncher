package utils

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
)

func SplitJSONToChunks(r io.Reader, chunkSize int) ([][]byte, error) {
	dec := json.NewDecoder(r)
	var chunks [][]byte
	var buffer []json.RawMessage

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('[') {
		return nil, io.ErrUnexpectedEOF
	}

	for dec.More() {
		var obj json.RawMessage
		if err := dec.Decode(&obj); err != nil {
			return nil, err
		}
		buffer = append(buffer, obj)

		if len(buffer) >= chunkSize {
			chunkBytes, _ := json.Marshal(buffer)
			chunks = append(chunks, chunkBytes)
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		chunkBytes, _ := json.Marshal(buffer)
		chunks = append(chunks, chunkBytes)
	}

	return chunks, nil
}

func SplitCSVToChunks(r io.Reader, chunkSize int) ([][]byte, error) {
	csvReader := csv.NewReader(r)
	var chunks [][]byte
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	count := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			if count > 0 {
				writer.Flush()
				chunks = append(chunks, buf.Bytes())
			}
			break
		}
		if err != nil {
			return nil, err
		}

		if err := writer.Write(record); err != nil {
			return nil, err
		}

		count++
		if count >= chunkSize {
			writer.Flush()
			chunks = append(chunks, buf.Bytes())
			buf.Reset()
			writer = csv.NewWriter(&buf)
			count = 0
		}
	}

	return chunks, nil
}
