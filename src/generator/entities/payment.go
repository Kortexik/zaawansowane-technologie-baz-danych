package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type PaymentGenerator struct {
	generator.Base
	payment types.Payment
}

func NewPaymentGenerator(id, orderID int) *PaymentGenerator {
	p := generator.GeneratePayment(id, orderID)
	return &PaymentGenerator{
		Base:    generator.NewBase("payments", "payments", "payment", p.ID, p),
		payment: p,
	}
}

func (g *PaymentGenerator) SQLInsert() string {
	p := g.payment
	return fmt.Sprintf(
		"INSERT INTO payments (id, order_id, amount, currency, method, status, transaction_id, created_at, provider, fee) VALUES (%d, %d, %.2f, '%s', '%s', '%s', '%s', '%s', '%s', %.2f);",
		p.ID, p.OrderID, p.Amount, p.Currency, p.Method, p.Status, p.TransactionID,
		p.CreatedAt.Format(generator.SQLTimeFormat), p.Provider, p.Fee,
	)
}

func (g *PaymentGenerator) MongoInsert() string {
	p := g.payment
	return fmt.Sprintf(
		"db.payments.insertOne({_id: %d, orderId: %d, amount: %.2f, currency: '%s', method: '%s', status: '%s', transactionId: '%s', createdAt: ISODate('%s'), provider: '%s', fee: %.2f});\n",
		p.ID, p.OrderID, p.Amount, p.Currency, p.Method, p.Status, p.TransactionID,
		p.CreatedAt.Format(time.RFC3339), p.Provider, p.Fee,
	)
}
