package baseMySQL

import (
	"fmt"
	"reflect"
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

// Scan implements the Scanner interface for NullTime
func (nt *NullTime) Scan(value interface{}) error {
	var t mysqlDrv.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nt = NullTime{t.Time, false}
	} else {
		*nt = NullTime{t.Time, true}
	}

	return nil
}
