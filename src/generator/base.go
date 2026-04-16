package generator

import (
	"encoding/json"
	"fmt"
)

// SQLTimeFormat is the datetime format used in SQL INSERT statements.
const SQLTimeFormat = "2006-01-02 15:04:05"

// Base holds the entity metadata and stored data needed to generate queries.
// Embed this in every entity generator to get all shared query methods for free.
type Base struct {
	table      string
	collection string
	keyPrefix  string
	id         int
	data       any // stored for Redis JSON serialisation
}

// NewBase constructs a Base with all required metadata.
func NewBase(table, collection, keyPrefix string, id int, data any) Base {
	return Base{table: table, collection: collection, keyPrefix: keyPrefix, id: id, data: data}
}

// BoolToInt converts a boolean to the 0/1 integer SQL expects.
// Deprecated: Use BoolToSQL for better compatibility.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// BoolToSQL converts a boolean to a SQL-compatible string.
// Returns "true"/"false" which works for both MySQL and PostgreSQL.
func BoolToSQL(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// FormatMongoValue wraps strings in single quotes; other types are printed as-is.
func FormatMongoValue(value any) string {
	if s, ok := value.(string); ok {
		return fmt.Sprintf("'%s'", s)
	}
	return fmt.Sprintf("%v", value)
}

// ----- SQL -----

func (b *Base) SQLSelect(id int) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE id = %d;", b.table, id)
}

func (b *Base) SQLUpdate(id int, field string, value any) string {
	return fmt.Sprintf("UPDATE %s SET %s = '%v' WHERE id = %d;", b.table, field, value, id)
}

func (b *Base) SQLDelete(id int) string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = %d;", b.table, id)
}

// ----- MongoDB -----

func (b *Base) MongoFind(id int) string {
	return fmt.Sprintf("db.%s.findOne({_id: %d});\n", b.collection, id)
}

func (b *Base) MongoUpdate(id int, field string, value any) string {
	return fmt.Sprintf("db.%s.updateOne({_id: %d}, {$set: {%s: %s}});\n",
		b.collection, id, field, FormatMongoValue(value))
}

func (b *Base) MongoDelete(id int) string {
	return fmt.Sprintf("db.%s.deleteOne({_id: %d});\n", b.collection, id)
}

// ----- Redis -----

func (b *Base) RedisSet() string {
	jsonData, _ := json.Marshal(b.data)
	return fmt.Sprintf("SET %s:%d %s\n", b.keyPrefix, b.id, string(jsonData))
}

func (b *Base) RedisGet(id int) string {
	return fmt.Sprintf("GET %s:%d\n", b.keyPrefix, id)
}

func (b *Base) RedisDelete(id int) string {
	return fmt.Sprintf("DEL %s:%d\n", b.keyPrefix, id)
}
