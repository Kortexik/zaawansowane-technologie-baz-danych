package entities

import (
	"benchmark/generator"
	"benchmark/types"
	"fmt"
	"time"
)

type InventoryLogGenerator struct {
	generator.Base
	log types.InventoryLog
}

func NewInventoryLogGenerator(id, productID int) *InventoryLogGenerator {
	l := generator.GenerateInventoryLog(id, productID)
	return &InventoryLogGenerator{
		Base: generator.NewBase("inventory_logs", "inventory_logs", "inventory_log", l.ID, l),
		log:  l,
	}
}

func (g *InventoryLogGenerator) SQLInsert() string {
	l := g.log
	return fmt.Sprintf(
		"INSERT INTO inventory_logs (id, product_id, change_type, quantity, old_quantity, new_quantity, reason, created_at, employee_id, warehouse_id) VALUES (%d, %d, '%s', %d, %d, %d, '%s', '%s', %d, %d);",
		l.ID, l.ProductID, l.ChangeType, l.Quantity, l.OldQuantity, l.NewQuantity, l.Reason,
		l.CreatedAt.Format(generator.SQLTimeFormat), l.EmployeeID, l.WarehouseID,
	)
}

func (g *InventoryLogGenerator) MongoInsert() string {
	l := g.log
	return fmt.Sprintf(
		"db.inventory_logs.insertOne({_id: %d, productId: %d, changeType: '%s', quantity: %d, oldQuantity: %d, newQuantity: %d, reason: '%s', createdAt: ISODate('%s'), employeeId: %d, warehouseId: %d});\n",
		l.ID, l.ProductID, l.ChangeType, l.Quantity, l.OldQuantity, l.NewQuantity, l.Reason,
		l.CreatedAt.Format(time.RFC3339), l.EmployeeID, l.WarehouseID,
	)
}
