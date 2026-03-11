package types

import (
	"time"
)

type User struct {
	ID        int
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	CreatedAt time.Time
	Country   string
	City      string
	IsActive  bool
}

type Address struct {
	ID         int
	UserID     int
	Street     string
	City       string
	PostalCode string
	Country    string
	CreatedAt  time.Time
	IsDefault  bool
	Latitude   float64
	Longitude  float64
}

type Category struct {
	ID           int
	Name         string
	Description  string
	ParentID     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsActive     bool
	Slug         string
	Icon         string
	DisplayOrder int
}

type Product struct {
	ID          int
	Name        string
	Description string
	CategoryID  int
	Price       float64
	Currency    string
	Stock       int
	Weight      float64
	CreatedAt   time.Time
	IsActive    bool
}

type ProductImage struct {
	ID        int
	ProductID int
	ImageURL  string
	AltText   string
	IsMain    bool
	CreatedAt time.Time
	Width     int
	Height    int
	Format    string
	SizeKB    int
}

type Order struct {
	ID            int
	UserID        int
	AddressID     int
	Status        string
	TotalAmount   float64
	Currency      string
	PaymentMethod string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ShippingCost  float64
}

type OrderItem struct {
	ID          int
	OrderID     int
	ProductID   int
	ProductName string
	Price       float64
	Quantity    int
	TotalPrice  float64
	CreatedAt   time.Time
	Discount    float64
	Tax         float64
}

type Payment struct {
	ID            int
	OrderID       int
	Amount        float64
	Currency      string
	Method        string
	Status        string
	TransactionID string
	CreatedAt     time.Time
	Provider      string
	Fee           float64
}

type Review struct {
	ID           int
	UserID       int
	ProductID    int
	Rating       int
	Title        string
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsVerified   bool
	HelpfulVotes int
}

type InventoryLog struct {
	ID          int
	ProductID   int
	ChangeType  string
	Quantity    int
	OldQuantity int
	NewQuantity int
	Reason      string
	CreatedAt   time.Time
	EmployeeID  int
	WarehouseID int
}
