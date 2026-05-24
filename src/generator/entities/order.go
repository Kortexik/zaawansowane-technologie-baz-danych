package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type OrderGenerator struct {
	generator.Base
	order types.Order
}

func NewOrderGenerator(id, usersTotal, addressesTotal int) *OrderGenerator {
	o := generator.GenerateOrder(id, usersTotal, addressesTotal)
	return &OrderGenerator{
		Base:  generator.NewBase("orders", "orders", "order", o.ID, o),
		order: o,
	}
}

func (g *OrderGenerator) SQLSelect(id int) string {
	return fmt.Sprintf("SELECT * FROM orders WHERE status = '%s' LIMIT 100;", g.order.Status)
}

func (g *OrderGenerator) MongoFind(id int) string {
	return fmt.Sprintf("db.orders.find({status: '%s'}).limit(100);\n", g.order.Status)
}

func (g *OrderGenerator) SQLInsert() string {
	o := g.order
	return fmt.Sprintf(
		"INSERT INTO orders (id, user_id, address_id, status, total_amount, currency, payment_method, created_at, updated_at, shipping_cost) VALUES (%d, %d, %d, '%s', %.2f, '%s', '%s', '%s', '%s', %.2f);",
		o.ID, o.UserID, o.AddressID, o.Status, o.TotalAmount, o.Currency, o.PaymentMethod,
		o.CreatedAt.Format(generator.SQLTimeFormat), o.UpdatedAt.Format(generator.SQLTimeFormat), o.ShippingCost,
	)
}

func (g *OrderGenerator) MongoInsert() string {
	o := g.order
	return fmt.Sprintf(
		"db.orders.insertOne({_id: %d, userId: %d, addressId: %d, status: '%s', totalAmount: %.2f, currency: '%s', paymentMethod: '%s', createdAt: ISODate('%s'), updatedAt: ISODate('%s'), shippingCost: %.2f});\n",
		o.ID, o.UserID, o.AddressID, o.Status, o.TotalAmount, o.Currency, o.PaymentMethod,
		o.CreatedAt.Format(time.RFC3339), o.UpdatedAt.Format(time.RFC3339), o.ShippingCost,
	)
}
