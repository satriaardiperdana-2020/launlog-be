package helper

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

func NumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	return float64(numericToFloat64(n))
}

// Helper function untuk konversi pgtype.Numeric ke float32
func NumericToFloat32(n pgtype.Numeric) float32 {
	if !n.Valid {
		return 0
	}
	return float32(numericToFloat64(n))
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	if n.Exp == 0 {
		f, _ := n.Int.Float64()
		return f
	}
	divisor := 1.0
	for i := n.Exp; i < 0; i++ {
		divisor *= 10
	}
	intVal, _ := n.Int.Float64()
	return intVal / divisor
}

// Float32ToNumeric mengkonversi float32 ke pgtype.Numeric
func Float32ToNumeric(f float32) pgtype.Numeric {
	n := pgtype.Numeric{}
	err := n.Scan(float64(f))
	if err != nil {
		n.Valid = false
	}
	return n
}

// Float64ToNumeric mengkonversi float64 ke pgtype.Numeric
func Float64ToNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	err := n.Scan(f)
	if err != nil {
		n.Valid = false
	}
	return n
}

// Helper function untuk konversi
func nullInt64ToInt64(n sql.NullInt64, defaultValue int64) int64 {
	if n.Valid {
		return n.Int64
	}
	return defaultValue
}
