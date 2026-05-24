package generator

import (
	"benchmark/types"
	"fmt"
	"math/rand"
	"time"
)

// Lists used by GenerateXxx — wider lists for indexed columns so Phase B
// SELECT-with-index queries get realistic per-match row counts (~hundreds,
// not the entire table). Without this the index on a low-cardinality column
// degenerates to a full scan because every row matches the predicate.
var (
	categoriesList = []string{"electronics", "accessories", "office", "gaming"}
	paymentsList   = []string{"credit_card", "paypal", "bank_transfer", "apple_pay"}
	productsList   = []string{"Laptop", "Phone", "Headphones", "Keyboard", "Mouse", "Monitor", "Camera", "Tablet", "Smartwatch", "Speaker"}
	currencies     = []string{"USD", "EUR", "PLN"}

	// 50 country names for users.country and addresses.country.
	countriesList = []string{
		"Poland", "Germany", "France", "Spain", "Italy", "United Kingdom", "Netherlands", "Belgium",
		"Austria", "Switzerland", "Sweden", "Norway", "Finland", "Denmark", "Czech Republic", "Slovakia",
		"Hungary", "Romania", "Bulgaria", "Greece", "Portugal", "Ireland", "Croatia", "Slovenia",
		"Estonia", "Latvia", "Lithuania", "USA", "Canada", "Mexico", "Brazil", "Argentina",
		"Chile", "Colombia", "Australia", "New Zealand", "Japan", "South Korea", "China", "India",
		"Indonesia", "Thailand", "Vietnam", "Philippines", "Turkey", "Israel", "Egypt", "Morocco",
		"South Africa", "Kenya",
	}

	// 100 city names for users.city and addresses.city.
	citiesList = []string{
		"Warsaw", "Krakow", "Berlin", "Munich", "Hamburg", "Paris", "Lyon", "Marseille",
		"Madrid", "Barcelona", "Rome", "Milan", "Naples", "London", "Manchester", "Liverpool",
		"Amsterdam", "Rotterdam", "Brussels", "Vienna", "Zurich", "Geneva", "Stockholm", "Oslo",
		"Helsinki", "Copenhagen", "Prague", "Bratislava", "Budapest", "Bucharest", "Sofia", "Athens",
		"Lisbon", "Porto", "Dublin", "Zagreb", "Ljubljana", "Tallinn", "Riga", "Vilnius",
		"New York", "Los Angeles", "Chicago", "Houston", "San Francisco", "Boston", "Seattle", "Miami",
		"Toronto", "Vancouver", "Montreal", "Mexico City", "Sao Paulo", "Rio de Janeiro", "Buenos Aires",
		"Santiago", "Bogota", "Lima", "Sydney", "Melbourne", "Auckland", "Wellington", "Tokyo",
		"Osaka", "Seoul", "Busan", "Shanghai", "Beijing", "Shenzhen", "Mumbai", "Delhi",
		"Bangalore", "Chennai", "Jakarta", "Bangkok", "Hanoi", "Ho Chi Minh City", "Manila", "Istanbul",
		"Ankara", "Tel Aviv", "Jerusalem", "Cairo", "Casablanca", "Cape Town", "Johannesburg", "Nairobi",
		"Lagos", "Riyadh", "Dubai", "Doha", "Kuwait City", "Muscat", "Beirut", "Amman",
		"Tehran", "Karachi", "Lahore", "Dhaka", "Colombo", "Kuala Lumpur",
	}

	// 200 order status values — extended past the standard 5 (pending/paid/shipped/delivered/returned)
	// so the orders.status index has selectivity comparable to other indexed columns
	// even at 10M rows (10M / 200 = 50k rows per match → ~0.5% selectivity, planner will use index).
	statusList = generateStatusList()

	// Fixed seed for reproducibility — every run produces the same data shape,
	// so comparisons across runs are meaningful. Was time.Now().UnixNano().
	rng = rand.New(rand.NewSource(42))
)

func generateStatusList() []string {
	bases := []string{
		"pending", "processing", "verified", "preparing", "packed",
		"ready", "shipped", "in_transit", "out_for_delivery", "delivered",
		"cancelled", "refunded", "returned", "disputed", "on_hold",
	}
	stages := []string{"step1", "step2", "step3", "step4", "step5",
		"step6", "step7", "step8", "step9", "step10",
		"step11", "step12", "step13", "step14"}
	out := make([]string, 0, len(bases)*len(stages))
	for _, b := range bases {
		for _, s := range stages {
			out = append(out, b+"_"+s)
		}
	}
	if len(out) > 200 {
		out = out[:200]
	}
	return out
}

func GenerateUser(id int) *types.User {
	return &types.User{
		ID:           id,
		Email:        fmt.Sprintf("user%d@example.com", id),
		PasswordHash: "hashed_pw",
		FirstName:    fmt.Sprintf("Name%d", id),
		LastName:     fmt.Sprintf("Surname%d", id),
		Phone:        fmt.Sprintf("+100000%04d", id%10000),
		CreatedAt:    time.Now(),
		Country:      countriesList[rng.Intn(len(countriesList))],
		City:         citiesList[rng.Intn(len(citiesList))],
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
		City:       citiesList[rng.Intn(len(citiesList))],
		PostalCode: fmt.Sprintf("%05d", rng.Intn(100000)),
		Country:    countriesList[rng.Intn(len(countriesList))],
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
