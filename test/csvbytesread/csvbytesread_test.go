package csvbytesread_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"testing"

	"github.com/eterline/geo-filt/pkg/csvbytesread"
)

func generateCSV(lines int) [][]string {
	data := make([][]string, lines)
	for i := 0; i < lines; i++ {
		data[i] = []string{
			fmt.Sprintf("192.168.%d.%d/32", i/256, i%256),
			"NA",
			"US",
			"\"Bonaire, Saint Eustatius and Saba\"",
		}
	}
	return data
}

func generateBytesCSV(lines int) (data []byte, n int) {
	records := generateCSV(lines)
	var buf bytes.Buffer

	writer := csv.NewWriter(&buf)
	writer.WriteAll(records)
	writer.Flush()
	return buf.Bytes(), len(records)
}

func BenchmarkCSVLoop(b *testing.B) {
	data, records := generateBytesCSV(256_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(data))
		r.ReuseRecord = true
		count := 0
		for {
			_, err := r.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			count++
		}
		if count != records {
			b.Fatalf("expected %d rows, got %d", records, count)
		}
	}
}

func BenchmarkCsvBytesReadLoop(b *testing.B) {
	data, records := generateBytesCSV(256_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := csvbytesread.NewByteCSVReader(bytes.NewReader(data))
		count := 0
		for {
			_, err := r.ReadRecord()
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
			count++
		}
		if count != records {
			b.Fatalf("expected %d rows, got %d", records, count)
		}
	}
}
