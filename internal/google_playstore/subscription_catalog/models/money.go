package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// -------------------------
// 🔹 MONEY STRUCT
// -------------------------

type Money struct {
	CurrencyCode string `gorm:"size:3" json:"currency_code"`
	Units        int64  `json:"units"`
	Nanos        int32  `json:"nanos"`
}

// ✅ Implement `sql.Valuer` to store Money as JSON in the database
func (m Money) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// ✅ Implement `sql.Scanner` to retrieve Money from JSON in the database
func (m *Money) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Money struct")
	}
	return json.Unmarshal(bytes, m)
}
