package csvbytesread

import (
	"bufio"
	"io"
)

type ByteCSVReader struct {
	r               *bufio.Reader
	FieldsPerRecord int
}

func NewByteCSVReader(r io.Reader) *ByteCSVReader {
	return &ByteCSVReader{
		r: bufio.NewReader(r),
	}
}

func (c *ByteCSVReader) ReadRecord() ([][]byte, error) {
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(line) == 0 {
			return nil, io.EOF
		}
	}

	return parseCSVLineBytes(line, ',')
}

func parseCSVLineBytes(line []byte, comma byte) ([][]byte, error) {
	var fields [][]byte
	inQuotes := false
	start := 0

	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inQuotes = !inQuotes
		case comma:
			if !inQuotes {
				fields = append(fields, line[start:i])
				start = i + 1
			}
		}
	}
	// последнее поле
	if start <= len(line) {
		fields = append(fields, line[start:])
	}
	return fields, nil
}
