package helper

import (
	"fmt"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	lastSeq map[string]int64
)

func init() {
	lastSeq = make(map[string]int64)
}

// GenerateInvoiceNo generates invoice number with format: INV-DDMMYYYY-0000000001
// Example: INV-20012026-0000000001
func GenerateInvoiceNo(prefix string) string {
	mu.Lock()
	defer mu.Unlock()

	// Get current date
	now := time.Now()
	dateStr := now.Format("02012006") // DDMMYYYY

	key := fmt.Sprintf("%s-%s", prefix, dateStr)

	// Increment sequence
	lastSeq[key]++
	seq := lastSeq[key]

	// Format sequence to 10 digits (0000000001)
	return fmt.Sprintf("%s-%s-%010d", prefix, dateStr, seq)
}
