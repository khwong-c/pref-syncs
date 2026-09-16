package sql

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/glebarez/sqlite"
	"github.com/samber/do/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JSONType json.RawMessage

var nullMsg JSONType

func init() {
	nullMsg, _ = json.RawMessage(nil).MarshalJSON()
}

func (*JSONType) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Name() {
	case "mysql", "sqlite":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

// Scan scans value into Jsonb, implements sql.Scanner interface.
func (j *JSONType) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	*j = bytes
	return nil
}

// Value returns json value, implement driver.Valuer interface.
func (j *JSONType) Value() (driver.Value, error) {
	if j == nil || len(*j) == 0 {
		return nullMsg, nil
	}
	return json.RawMessage(*j), nil
}

// NewSQLite constructs a file-based SQLite gorm.DB connection.
func NewSQLite(file string) (*gorm.DB, error) {
	return gorm.Open(
		sqlite.Open(file),
		&gorm.Config{
			PrepareStmt: true,
		},
	)
}

// NewInMemorySQLite constructs an in-memory SQLite gorm.DB connection.
func NewInMemorySQLite(do.Injector) (*gorm.DB, error) {
	return gorm.Open(
		sqlite.Open(generateInMemDatabaseDSN()),
		&gorm.Config{
			PrepareStmt: true,
		},
	)
}

func generateInMemDatabaseDSN() string {
	const maxRange = 1 << 62
	rndInt, err := rand.Int(rand.Reader, big.NewInt(maxRange))
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		"file:/tmp/dnd.%s.db?mode=memory&cache=shared",
		rndInt.String(),
	)
}
