# Analiza planów zapytań (EXPLAIN)

Plany zebrane dla pierwszego zapytania z każdego pliku SELECT,
przy największej dostępnej skali danych. Pokazują wpływ utworzenia
indeksu na wybór ścieżki dostępu przez optymalizator.

---

## MYSQL · `addresses`

**Zapytanie**: `SELECT * FROM addresses WHERE postal_code = '63408' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | addresses | NULL | ALL | NULL | NULL | NULL | NULL | 497160 | 10 | Using where
```

### Z indeksem

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | addresses | NULL | ref | idx_addresses_postal_code | idx_addresses_postal_code | 83 | const | 104 | 100 | NULL
```

---

## MYSQL · `categories`

**Zapytanie**: `SELECT * FROM categories WHERE display_order = 1 LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | categories | NULL | ALL | NULL | NULL | NULL | NULL | 200 | 10 | Using where
```

### Z indeksem

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | categories | NULL | ref | idx_categories_display_order | idx_categories_display_order | 5 | const | 1 | 100 | NULL
```

---

## MYSQL · `orders`

**Zapytanie**: `SELECT * FROM orders WHERE status = 'ready_step2' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | orders | NULL | ALL | NULL | NULL | NULL | NULL | 497344 | 10 | Using where
```

### Z indeksem

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | orders | NULL | ref | idx_orders_status | idx_orders_status | 203 | const | 93440 | 100 | NULL
```

---

## MYSQL · `products`

**Zapytanie**: `SELECT * FROM products WHERE category_id = 132 LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | products | NULL | ALL | NULL | NULL | NULL | NULL | 99699 | 10 | Using where
```

### Z indeksem

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | products | NULL | ref | idx_products_category_id | idx_products_category_id | 9 | const | 487 | 100 | NULL
```

---

## MYSQL · `users`

**Zapytanie**: `SELECT * FROM users WHERE country = 'Colombia' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | users | NULL | ALL | NULL | NULL | NULL | NULL | 496097 | 10 | Using where
```

### Z indeksem

```
id | select_type | table | partitions | type | possible_keys | key | key_len | ref | rows | filtered | Extra
------------------------------------------------------------------------------------------------------------
1 | SIMPLE | users | NULL | ref | idx_users_country | idx_users_country | 403 | const | 399974 | 100 | NULL
```

---

## POSTGRESQL · `addresses`

**Zapytanie**: `SELECT * FROM addresses WHERE postal_code = '63408' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
Limit  (cost=1000.00..10724.77 rows=6 width=75)
  ->  Gather  (cost=1000.00..10724.77 rows=6 width=75)
        Workers Planned: 2
        ->  Parallel Seq Scan on addresses  (cost=0.00..9724.17 rows=2 width=75)
              Filter: ((postal_code)::text = '63408'::text)
```

### Z indeksem

```
Limit  (cost=5.22..398.48 rows=100 width=75)
  ->  Bitmap Heap Scan on addresses  (cost=5.22..402.41 rows=101 width=75)
        Recheck Cond: ((postal_code)::text = '63408'::text)
        ->  Bitmap Index Scan on idx_addresses_postal_code  (cost=0.00..5.19 rows=101 width=0)
              Index Cond: ((postal_code)::text = '63408'::text)
```

---

## POSTGRESQL · `categories`

**Zapytanie**: `SELECT * FROM categories WHERE display_order = 1 LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
Limit  (cost=0.00..6.50 rows=1 width=95)
  ->  Seq Scan on categories  (cost=0.00..6.50 rows=1 width=95)
        Filter: (display_order = 1)
```

### Z indeksem

```
Limit  (cost=0.00..6.50 rows=1 width=95)
  ->  Seq Scan on categories  (cost=0.00..6.50 rows=1 width=95)
        Filter: (display_order = 1)
```

---

## POSTGRESQL · `orders`

**Zapytanie**: `SELECT * FROM orders WHERE status = 'ready_step2' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
Limit  (cost=0.00..535.95 rows=100 width=79)
  ->  Seq Scan on orders  (cost=0.00..13313.03 rows=2484 width=79)
        Filter: ((status)::text = 'ready_step2'::text)
```

### Z indeksem

```
Limit  (cost=0.43..342.61 rows=100 width=79)
  ->  Index Scan using idx_orders_status on orders  (cost=0.43..170699.15 rows=49886 width=79)
        Index Cond: ((status)::text = 'ready_step2'::text)
```

---

## POSTGRESQL · `products`

**Zapytanie**: `SELECT * FROM products WHERE category_id = 132 LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
Limit  (cost=0.00..542.51 rows=100 width=74)
  ->  Seq Scan on products  (cost=0.00..2680.00 rows=494 width=74)
        Filter: (category_id = 132)
```

### Z indeksem

```
Limit  (cost=8.13..211.45 rows=100 width=72)
  ->  Bitmap Heap Scan on products  (cost=8.13..1014.58 rows=495 width=72)
        Recheck Cond: (category_id = 132)
        ->  Bitmap Index Scan on idx_products_category_id  (cost=0.00..8.00 rows=495 width=0)
              Index Cond: (category_id = 132)
```

---

## POSTGRESQL · `users`

**Zapytanie**: `SELECT * FROM users WHERE country = 'Colombia' LIMIT 100`  
**Skala**: 500000 rows

### Bez indeksu

```
Limit  (cost=0.00..138.25 rows=100 width=100)
  ->  Seq Scan on users  (cost=0.00..14700.00 rows=10633 width=100)
        Filter: ((country)::text = 'Colombia'::text)
```

### Z indeksem

```
Limit  (cost=0.00..148.16 rows=100 width=103)
  ->  Seq Scan on users  (cost=0.00..302807.79 rows=204377 width=103)
        Filter: ((country)::text = 'Colombia'::text)
```

---

