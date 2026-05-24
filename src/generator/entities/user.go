package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type UserGenerator struct {
	generator.Base
	user *types.User
}

func NewUserGenerator(id int) *UserGenerator {
	u := generator.GenerateUser(id)
	return &UserGenerator{
		Base: generator.NewBase("users", "users", "user", u.ID, u),
		user: u,
	}
}

func (g *UserGenerator) SQLSelect(id int) string {
	return fmt.Sprintf("SELECT * FROM users WHERE country = '%s' LIMIT 100;", g.user.Country)
}

// MongoFind overrides the PK-based base so the Mongo benchmark hits the same
// indexed column as the SQL benchmark — apples-to-apples comparison.
func (g *UserGenerator) MongoFind(id int) string {
	return fmt.Sprintf("db.users.find({country: '%s'}).limit(100);\n", g.user.Country)
}

func (g *UserGenerator) SQLInsert() string {
	u := g.user
	return fmt.Sprintf(
		"INSERT INTO users (id, email, password_hash, first_name, last_name, phone, created_at, country, city, is_active) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %s);",
		u.ID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone,
		u.CreatedAt.Format(generator.SQLTimeFormat), u.Country, u.City, generator.BoolToSQL(u.IsActive),
	)
}

func (g *UserGenerator) MongoInsert() string {
	u := g.user
	return fmt.Sprintf(
		"db.users.insertOne({_id: %d, email: '%s', passwordHash: '%s', firstName: '%s', lastName: '%s', phone: '%s', createdAt: ISODate('%s'), country: '%s', city: '%s', isActive: %v});\n",
		u.ID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone,
		u.CreatedAt.Format(time.RFC3339), u.Country, u.City, u.IsActive,
	)
}
