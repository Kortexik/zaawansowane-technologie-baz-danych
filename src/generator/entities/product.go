package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type ProductGenerator struct {
	generator.Base
	product types.Product
}

func NewProductGenerator(id int) *ProductGenerator {
	p := generator.GenerateProduct(id)
	return &ProductGenerator{
		Base:    generator.NewBase("products", "products", "product", p.ID, p),
		product: p,
	}
}

func (g *ProductGenerator) SQLInsert() string {
	p := g.product
	return fmt.Sprintf(
		"INSERT INTO products (id, name, description, category_id, price, currency, stock, weight, created_at, is_active) VALUES (%d, '%s', '%s', %d, %.2f, '%s', %d, %.2f, '%s', 1);",
		p.ID, p.Name, p.Description, p.CategoryID, p.Price, p.Currency, p.Stock, p.Weight,
		p.CreatedAt.Format(generator.SQLTimeFormat),
	)
}

func (g *ProductGenerator) MongoInsert() string {
	p := g.product
	return fmt.Sprintf(
		"db.products.insertOne({_id: %d, name: '%s', description: '%s', categoryId: %d, price: %.2f, currency: '%s', stock: %d, weight: %.2f, createdAt: ISODate('%s'), isActive: %v});\n",
		p.ID, p.Name, p.Description, p.CategoryID, p.Price, p.Currency, p.Stock, p.Weight,
		p.CreatedAt.Format(time.RFC3339), p.IsActive,
	)
}
