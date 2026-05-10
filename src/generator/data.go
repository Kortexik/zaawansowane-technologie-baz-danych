package generator

import (
	"benchmark/types"
	"fmt"
	"math/rand"
	"time"
)

var (
	categoriesList = []string{"electronics", "accessories", "office", "gaming"}
	paymentsList   = []string{"credit_card", "paypal", "bank_transfer", "apple_pay"}
	productsList   = []string{"Laptop", "Phone", "Headphones", "Keyboard", "Mouse", "Monitor", "Camera", "Tablet", "Smartwatch", "Speaker"}
	statusList     = []string{"pending", "paid", "shipped", "delivered", "returned"}
	currencies     = []string{"USD", "EUR", "PLN"}
	rng            = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GenerateUser(id int) *types.User {
	return &types.User{
		ID:           id,
		Email:        fmt.Sprintf("user%d@example.com", id),
		PasswordHash: "hashed_pw",
		FirstName:    fmt.Sprintf("Name%d", id),
		LastName:     fmt.Sprintf("Surname%d", id),
		Phone:        fmt.Sprintf("+100000%04d", id%10000),
		CreatedAt:    time.Now(),
		Country:      "Poland",
		City:         "Warsaw",
		IsActive:     rng.Intn(2) == 0,
	}
}

func GenerateProduct(id int) types.Product {
	name := productsList[rng.Intn(len(productsList))]
	return types.Product{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("%s description", name),
		CategoryID:  rng.Intn(200) + 1,
		Price:       float64(rng.Intn(2000) + 10),
		Currency:    currencies[rng.Intn(len(currencies))],
		Stock:       rng.Intn(1000),
		Weight:      float64(rng.Intn(1000))/100.0 + 0.1,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
}

func GenerateAddress(id, userID int) types.Address {
	return types.Address{
		ID:         id,
		UserID:     userID,
		Street:     fmt.Sprintf("%d Main St", rng.Intn(9999)+1),
		City:       "Warsaw",
		PostalCode: fmt.Sprintf("%05d", rng.Intn(100000)),
		Country:    "Poland",
		CreatedAt:  time.Now(),
		IsDefault:  rng.Intn(2) == 0,
		Latitude:   52.2297 + (rng.Float64()-0.5)*0.1,
		Longitude:  21.0122 + (rng.Float64()-0.5)*0.1,
	}
}

func GenerateCategory(id int) types.Category {
	return types.Category{
		ID:           id,
		Name:         fmt.Sprintf("Category%d", id),
		Description:  fmt.Sprintf("Description for category %d", id),
		ParentID:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
		Slug:         fmt.Sprintf("category-%d", id),
		Icon:         fmt.Sprintf("icon%d", id),
		DisplayOrder: id,
	}
}

func GenerateProductImage(id, productID int) types.ProductImage {
	return types.ProductImage{
		ID:        id,
		ProductID: productID,
		ImageURL:  fmt.Sprintf("https://example.com/images/product-%d-%d.jpg", productID, id),
		AltText:   fmt.Sprintf("Product image %d", id),
		IsMain:    id == 1,
		CreatedAt: time.Now(),
		Width:     1920,
		Height:    1080,
		Format:    "jpg",
		SizeKB:    rng.Intn(2000) + 500,
	}
}

func GenerateOrder(id, usersTotal, addressesTotal int) types.Order {
	return types.Order{
		ID:            id,
		UserID:        rng.Intn(usersTotal) + 1,
		AddressID:     rng.Intn(addressesTotal) + 1,
		Status:        statusList[rng.Intn(len(statusList))],
		TotalAmount:   float64(rng.Intn(2000) + 20),
		Currency:      currencies[rng.Intn(len(currencies))],
		PaymentMethod: paymentsList[rng.Intn(len(paymentsList))],
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ShippingCost:  float64(rng.Intn(50) + 5),
	}
}

func GenerateOrderItem(id, orderID, productID int) types.OrderItem {
	price := float64(rng.Intn(2000) + 10)
	quantity := rng.Intn(10) + 1
	discount := float64(rng.Intn(50)) / 100
	subtotal := price * float64(quantity)
	totalPrice := subtotal - subtotal*discount
	return types.OrderItem{
		ID:          id,
		OrderID:     orderID,
		ProductID:   productID,
		ProductName: fmt.Sprintf("Product%d", productID),
		Price:       price,
		Quantity:    quantity,
		TotalPrice:  totalPrice,
		CreatedAt:   time.Now(),
		Discount:    discount,
		Tax:         totalPrice * 0.23,
	}
}

func GeneratePayment(id, orderID int) types.Payment {
	amount := float64(rng.Intn(2000) + 20)
	return types.Payment{
		ID:            id,
		OrderID:       orderID,
		Amount:        amount,
		Currency:      currencies[rng.Intn(len(currencies))],
		Method:        paymentsList[rng.Intn(len(paymentsList))],
		Status:        statusList[rng.Intn(len(statusList))],
		TransactionID: fmt.Sprintf("TXN%d", id),
		CreatedAt:     time.Now(),
		Provider:      fmt.Sprintf("Provider%d", rng.Intn(5)+1),
		Fee:           amount * 0.03,
	}
}

func GenerateReview(id, userID, productID int) types.Review {
	return types.Review{
		ID:           id,
		UserID:       userID,
		ProductID:    productID,
		Rating:       rng.Intn(5) + 1,
		Title:        fmt.Sprintf("Review%d", id),
		Comment:      fmt.Sprintf("This is a review comment for review %d", id),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsVerified:   rng.Intn(2) == 0,
		HelpfulVotes: rng.Intn(100),
	}
}

func GenerateInventoryLog(id, productID int) types.InventoryLog {
	oldQty := rng.Intn(1000)
	newQty := oldQty + rng.Intn(100) - rng.Intn(100)
	return types.InventoryLog{
		ID:          id,
		ProductID:   productID,
		ChangeType:  []string{"in", "out", "adjustment"}[rng.Intn(3)],
		Quantity:    newQty - oldQty,
		OldQuantity: oldQty,
		NewQuantity: newQty,
		Reason:      []string{"purchase", "restock", "damage", "return"}[rng.Intn(4)],
		CreatedAt:   time.Now(),
		EmployeeID:  rng.Intn(100) + 1,
		WarehouseID: rng.Intn(5) + 1,
	}
}
