package mysql

import (
	"fmt"
	"time"

	mysqlDrv "github.com/go-sql-driver/mysql"
)

// NullTime is wrapper for mysql.NullTime
// Implements json.Marshaler and sql.Scanner
type NullTime mysqlDrv.NullTime

// MarshalJSON for NullTime
func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
	return []byte(val), nil
}

