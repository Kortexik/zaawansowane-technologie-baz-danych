-- MySQL schema. Column lists mirror what src/generator emits in the INSERT
-- statements written to queries/*_sql_insert.sql — keep these two in sync.

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    created_at TIMESTAMP NULL,
    country VARCHAR(100),
    city VARCHAR(100),
    is_active BOOLEAN
);

CREATE TABLE IF NOT EXISTS addresses (
    id BIGINT PRIMARY KEY,
    user_id BIGINT,
    street VARCHAR(255),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    created_at TIMESTAMP NULL,
    is_default BOOLEAN,
    latitude DOUBLE,
    longitude DOUBLE
);

CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    parent_id BIGINT,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    is_active BOOLEAN,
    slug VARCHAR(255),
    icon VARCHAR(255),
    display_order INT
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    category_id BIGINT,
    price DECIMAL(10,2),
    currency VARCHAR(10),
    stock_quantity INT,
    weight DOUBLE,
    created_at TIMESTAMP NULL,
    is_active BOOLEAN
);

CREATE TABLE IF NOT EXISTS product_images (
    id BIGINT PRIMARY KEY,
    product_id BIGINT,
    image_url TEXT,
    alt_text VARCHAR(255),
    is_main BOOLEAN,
    created_at TIMESTAMP NULL,
    width INT,
    height INT,
    format VARCHAR(20),
    size_kb INT
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT,
    address_id BIGINT,
    status VARCHAR(50),
    total_amount DECIMAL(12,2),
    currency VARCHAR(10),
    payment_method VARCHAR(50),
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    shipping_cost DECIMAL(10,2)
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT PRIMARY KEY,
    order_id BIGINT,
    product_id BIGINT,
    product_name VARCHAR(255),
    price DECIMAL(10,2),
    quantity INT,
    total_price DECIMAL(12,2),
    created_at TIMESTAMP NULL,
    discount DECIMAL(10,2),
    tax DECIMAL(10,2)
);

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT PRIMARY KEY,
    order_id BIGINT,
    amount DECIMAL(12,2),
    currency VARCHAR(10),
    payment_method VARCHAR(50),
    payment_status VARCHAR(50),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP NULL,
    provider VARCHAR(100),
    fee DECIMAL(10,2)
);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGINT PRIMARY KEY,
    user_id BIGINT,
    product_id BIGINT,
    rating INT,
    title VARCHAR(255),
    comment TEXT,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    is_verified BOOLEAN,
    helpful_votes INT
);

CREATE TABLE IF NOT EXISTS inventory_logs (
    id BIGINT PRIMARY KEY,
    product_id BIGINT,
    change_type VARCHAR(50),
    quantity_change INT,
    old_quantity INT,
    new_quantity INT,
    reason VARCHAR(255),
    created_at TIMESTAMP NULL,
    employee_id BIGINT,
    warehouse_id BIGINT
);
