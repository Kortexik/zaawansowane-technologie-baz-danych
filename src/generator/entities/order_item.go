package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type OrderItemGenerator struct {
	generator.Base
	item types.OrderItem
}

func NewOrderItemGenerator(id, orderID, productID int) *OrderItemGenerator {
	oi := generator.GenerateOrderItem(id, orderID, productID)
	return &OrderItemGenerator{
		Base: generator.NewBase("order_items", "order_items", "order_item", oi.ID, oi),
		item: oi,
	}
}

func (g *OrderItemGenerator) SQLInsert() string {
	oi := g.item
	return fmt.Sprintf(
		"INSERT INTO order_items (id, order_id, product_id, product_name, price, quantity, total_price, created_at, discount, tax) VALUES (%d, %d, %d, '%s', %.2f, %d, %.2f, '%s', %.2f, %.2f);",
		oi.ID, oi.OrderID, oi.ProductID, oi.ProductName, oi.Price, oi.Quantity, oi.TotalPrice,
		oi.CreatedAt.Format(generator.SQLTimeFormat), oi.Discount, oi.Tax,
	)
}

func (g *OrderItemGenerator) MongoInsert() string {
	oi := g.item
	return fmt.Sprintf(
		"db.order_items.insertOne({_id: %d, orderId: %d, productId: %d, productName: '%s', price: %.2f, quantity: %d, totalPrice: %.2f, createdAt: ISODate('%s'), discount: %.2f, tax: %.2f});\n",
		oi.ID, oi.OrderID, oi.ProductID, oi.ProductName, oi.Price, oi.Quantity, oi.TotalPrice,
		oi.CreatedAt.Format(time.RFC3339), oi.Discount, oi.Tax,
	)
}
