package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type ProductImageGenerator struct {
	generator.Base
	image types.ProductImage
}

func NewProductImageGenerator(id, productID int) *ProductImageGenerator {
	img := generator.GenerateProductImage(id, productID)
	return &ProductImageGenerator{
		Base:  generator.NewBase("product_images", "product_images", "product_image", img.ID, img),
		image: img,
	}
}

func (g *ProductImageGenerator) SQLInsert() string {
	pi := g.image
	return fmt.Sprintf(
		"INSERT INTO product_images (id, product_id, image_url, alt_text, is_main, created_at, width, height, format, size_kb) VALUES (%d, %d, '%s', '%s', %d, '%s', %d, %d, '%s', %d);",
		pi.ID, pi.ProductID, pi.ImageURL, pi.AltText, generator.BoolToInt(pi.IsMain),
		pi.CreatedAt.Format(generator.SQLTimeFormat), pi.Width, pi.Height, pi.Format, pi.SizeKB,
	)
}

func (g *ProductImageGenerator) MongoInsert() string {
	pi := g.image
	return fmt.Sprintf(
		"db.product_images.insertOne({_id: %d, productId: %d, imageUrl: '%s', altText: '%s', isMain: %v, createdAt: ISODate('%s'), width: %d, height: %d, format: '%s', sizeKb: %d});\n",
		pi.ID, pi.ProductID, pi.ImageURL, pi.AltText, pi.IsMain,
		pi.CreatedAt.Format(time.RFC3339), pi.Width, pi.Height, pi.Format, pi.SizeKB,
	)
}
