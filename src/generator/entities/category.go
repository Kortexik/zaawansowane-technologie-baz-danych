package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type CategoryGenerator struct {
	generator.Base
	category types.Category
}

func NewCategoryGenerator(id int) *CategoryGenerator {
	c := generator.GenerateCategory(id)
	return &CategoryGenerator{
		Base:     generator.NewBase("categories", "categories", "category", c.ID, c),
		category: c,
	}
}

// SQLSelect overrides Base.SQLSelect to filter on display_order, the column
// we explicitly index in CreateIndexes(). Filtering on PK would always hit
// the implicit primary key index and hide the effect.
func (g *CategoryGenerator) SQLSelect(id int) string {
	return fmt.Sprintf("SELECT * FROM categories WHERE display_order = %d;", g.category.DisplayOrder)
}

func (g *CategoryGenerator) SQLInsert() string {
	c := g.category
	return fmt.Sprintf(
		"INSERT INTO categories (id, name, description, parent_id, created_at, updated_at, is_active, slug, icon, display_order) VALUES (%d, '%s', '%s', %d, '%s', '%s', %s, '%s', '%s', %d);",
		c.ID, c.Name, c.Description, c.ParentID,
		c.CreatedAt.Format(generator.SQLTimeFormat), c.UpdatedAt.Format(generator.SQLTimeFormat),
		generator.BoolToSQL(c.IsActive), c.Slug, c.Icon, c.DisplayOrder,
	)
}

func (g *CategoryGenerator) MongoInsert() string {
	c := g.category
	return fmt.Sprintf(
		"db.categories.insertOne({_id: %d, name: '%s', description: '%s', parentId: %d, createdAt: ISODate('%s'), updatedAt: ISODate('%s'), isActive: %v, slug: '%s', icon: '%s', displayOrder: %d});\n",
		c.ID, c.Name, c.Description, c.ParentID,
		c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339),
		c.IsActive, c.Slug, c.Icon, c.DisplayOrder,
	)
}
