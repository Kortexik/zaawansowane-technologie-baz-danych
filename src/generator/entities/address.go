package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type AddressGenerator struct {
	generator.Base
	address types.Address
}

func NewAddressGenerator(id, userID int) *AddressGenerator {
	a := generator.GenerateAddress(id, userID)
	return &AddressGenerator{
		Base:    generator.NewBase("addresses", "addresses", "address", a.ID, a),
		address: a,
	}
}

func (g *AddressGenerator) SQLInsert() string {
	a := g.address
	return fmt.Sprintf(
		"INSERT INTO addresses (id, user_id, street, city, postal_code, country, created_at, is_default, latitude, longitude) VALUES (%d, %d, '%s', '%s', '%s', '%s', '%s', %s, %.6f, %.6f);",
		a.ID, a.UserID, a.Street, a.City, a.PostalCode, a.Country,
		a.CreatedAt.Format(generator.SQLTimeFormat), generator.BoolToSQL(a.IsDefault), a.Latitude, a.Longitude,
	)
}

func (g *AddressGenerator) MongoInsert() string {
	a := g.address
	return fmt.Sprintf(
		"db.addresses.insertOne({_id: %d, userId: %d, street: '%s', city: '%s', postalCode: '%s', country: '%s', createdAt: ISODate('%s'), isDefault: %v, location: {type: 'Point', coordinates: [%.6f, %.6f]}});\n",
		a.ID, a.UserID, a.Street, a.City, a.PostalCode, a.Country,
		a.CreatedAt.Format(time.RFC3339), a.IsDefault, a.Longitude, a.Latitude,
	)
}
