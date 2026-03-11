package main

import (
	"benchmark/generator/entities"
	"benchmark/writer"
	"fmt"
	"math/rand"
	"os"
)

const (
	batchSize = 100_000

	numUsers      = 1_000_000
	numCategories = 200
	numProducts   = 100_000

	numAddresses     = numUsers
	numOrders        = numUsers
	numOrderItems    = numOrders
	numPayments      = numOrders
	numReviews       = numUsers / 5
	numProductImages = numProducts
	numInventoryLogs = numProducts
)

func main() {
	fmt.Println("Starting benchmark data generation...")

	type job struct {
		file  string
		label string
		count int
		query func(i int) string
	}

	jobs := []job{
		// ── Categories ──────────────────────────────────────────────────────────
		{"categories_sql_insert.sql", "category SQL inserts", numCategories, func(i int) string {
			return entities.NewCategoryGenerator(i).SQLInsert()
		}},
		{"categories_mongo_insert.js", "category Mongo inserts", numCategories, func(i int) string {
			return entities.NewCategoryGenerator(i).MongoInsert()
		}},
		{"categories_redis_set.txt", "category Redis sets", numCategories, func(i int) string {
			return entities.NewCategoryGenerator(i).RedisSet()
		}},

		// ── Users ────────────────────────────────────────────────────────────────
		{"users_sql_insert.sql", "user SQL inserts", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).SQLInsert()
		}},
		{"users_sql_select.sql", "user SQL selects", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).SQLSelect(i)
		}},
		{"users_sql_update.sql", "user SQL updates", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).SQLUpdate(i, "email", "updated@example.com")
		}},
		{"users_sql_delete.sql", "user SQL deletes", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).SQLDelete(i)
		}},
		{"users_mongo_insert.js", "user Mongo inserts", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).MongoInsert()
		}},
		{"users_mongo_find.js", "user Mongo finds", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).MongoFind(i)
		}},
		{"users_mongo_update.js", "user Mongo updates", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).MongoUpdate(i, "email", "updated@example.com")
		}},
		{"users_mongo_delete.js", "user Mongo deletes", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).MongoDelete(i)
		}},
		{"users_redis_set.txt", "user Redis sets", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).RedisSet()
		}},
		{"users_redis_get.txt", "user Redis gets", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).RedisGet(i)
		}},
		{"users_redis_delete.txt", "user Redis deletes", numUsers, func(i int) string {
			return entities.NewUserGenerator(i).RedisDelete(i)
		}},

		// ── Addresses ────────────────────────────────────────────────────────────
		{"addresses_sql_insert.sql", "address SQL inserts", numAddresses, func(i int) string {
			return entities.NewAddressGenerator(i, i).SQLInsert()
		}},
		{"addresses_mongo_insert.js", "address Mongo inserts", numAddresses, func(i int) string {
			return entities.NewAddressGenerator(i, i).MongoInsert()
		}},
		{"addresses_redis_set.txt", "address Redis sets", numAddresses, func(i int) string {
			return entities.NewAddressGenerator(i, i).RedisSet()
		}},

		// ── Products ─────────────────────────────────────────────────────────────
		{"products_sql_insert.sql", "product SQL inserts", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).SQLInsert()
		}},
		{"products_sql_select.sql", "product SQL selects", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).SQLSelect(i)
		}},
		{"products_sql_update.sql", "product SQL updates", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).SQLUpdate(i, "price", 99.99)
		}},
		{"products_sql_delete.sql", "product SQL deletes", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).SQLDelete(i)
		}},
		{"products_mongo_insert.js", "product Mongo inserts", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).MongoInsert()
		}},
		{"products_mongo_find.js", "product Mongo finds", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).MongoFind(i)
		}},
		{"products_mongo_update.js", "product Mongo updates", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).MongoUpdate(i, "price", 99.99)
		}},
		{"products_mongo_delete.js", "product Mongo deletes", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).MongoDelete(i)
		}},
		{"products_redis_set.txt", "product Redis sets", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).RedisSet()
		}},
		{"products_redis_get.txt", "product Redis gets", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).RedisGet(i)
		}},
		{"products_redis_delete.txt", "product Redis deletes", numProducts, func(i int) string {
			return entities.NewProductGenerator(i).RedisDelete(i)
		}},

		// ── Product Images ───────────────────────────────────────────────────────
		{"product_images_sql_insert.sql", "product image SQL inserts", numProductImages, func(i int) string {
			return entities.NewProductImageGenerator(i, i).SQLInsert()
		}},
		{"product_images_mongo_insert.js", "product image Mongo inserts", numProductImages, func(i int) string {
			return entities.NewProductImageGenerator(i, i).MongoInsert()
		}},

		// ── Orders ───────────────────────────────────────────────────────────────
		{"orders_sql_insert.sql", "order SQL inserts", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).SQLInsert()
		}},
		{"orders_sql_select.sql", "order SQL selects", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).SQLSelect(i)
		}},
		{"orders_sql_update.sql", "order SQL updates", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).SQLUpdate(i, "status", "shipped")
		}},
		{"orders_sql_delete.sql", "order SQL deletes", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).SQLDelete(i)
		}},
		{"orders_mongo_insert.js", "order Mongo inserts", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).MongoInsert()
		}},
		{"orders_mongo_find.js", "order Mongo finds", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).MongoFind(i)
		}},
		{"orders_redis_set.txt", "order Redis sets", numOrders, func(i int) string {
			return entities.NewOrderGenerator(i, numUsers, numAddresses).RedisSet()
		}},

		// ── Order Items ──────────────────────────────────────────────────────────
		{"order_items_sql_insert.sql", "order item SQL inserts", numOrderItems, func(i int) string {
			return entities.NewOrderItemGenerator(i, i, rand.Intn(numProducts)+1).SQLInsert()
		}},
		{"order_items_mongo_insert.js", "order item Mongo inserts", numOrderItems, func(i int) string {
			return entities.NewOrderItemGenerator(i, i, rand.Intn(numProducts)+1).MongoInsert()
		}},

		// ── Payments ─────────────────────────────────────────────────────────────
		{"payments_sql_insert.sql", "payment SQL inserts", numPayments, func(i int) string {
			return entities.NewPaymentGenerator(i, i).SQLInsert()
		}},
		{"payments_mongo_insert.js", "payment Mongo inserts", numPayments, func(i int) string {
			return entities.NewPaymentGenerator(i, i).MongoInsert()
		}},
		{"payments_redis_set.txt", "payment Redis sets", numPayments, func(i int) string {
			return entities.NewPaymentGenerator(i, i).RedisSet()
		}},

		// ── Reviews ──────────────────────────────────────────────────────────────
		{"reviews_sql_insert.sql", "review SQL inserts", numReviews, func(i int) string {
			return entities.NewReviewGenerator(i, rand.Intn(numUsers)+1, rand.Intn(numProducts)+1).SQLInsert()
		}},
		{"reviews_mongo_insert.js", "review Mongo inserts", numReviews, func(i int) string {
			return entities.NewReviewGenerator(i, rand.Intn(numUsers)+1, rand.Intn(numProducts)+1).MongoInsert()
		}},

		// ── Inventory Logs ───────────────────────────────────────────────────────
		{"inventory_logs_sql_insert.sql", "inventory log SQL inserts", numInventoryLogs, func(i int) string {
			return entities.NewInventoryLogGenerator(i, rand.Intn(numProducts)+1).SQLInsert()
		}},
		{"inventory_logs_mongo_insert.js", "inventory log Mongo inserts", numInventoryLogs, func(i int) string {
			return entities.NewInventoryLogGenerator(i, rand.Intn(numProducts)+1).MongoInsert()
		}},
	}

	for _, j := range jobs {
		fmt.Printf("\n--- %s ---\n", j.label)
		path := fmt.Sprintf("../queries/%s", j.file)
		f := mustCreate(path)
		generateQueries(f, j.label, j.count, j.query)
	}

	fmt.Println("\nBenchmark data generation complete!")
}

func generateQueries(file *os.File, label string, count int, fn func(i int) string) {
	bw := writer.New(file, batchSize)
	defer bw.Close()

	for i := 1; i <= count; i++ {
		bw.Write(fn(i))
		if i%batchSize == 0 {
			fmt.Printf("  %d / %d %s\n", i, count, label)
		}
	}
}

func mustCreate(path string) *os.File {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	return f
}
