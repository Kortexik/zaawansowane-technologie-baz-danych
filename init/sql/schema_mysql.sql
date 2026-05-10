-- MySQL Schema
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    created_at TIMESTAMP,
    country VARCHAR(100),
    city VARCHAR(100),
    is_active BOOLEAN
);

CREATE TABLE IF NOT EXISTS addresses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    street VARCHAR(255),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    created_at TIMESTAMP,
    is_default BOOLEAN,
    latitude DOUBLE,
    longitude DOUBLE
);

CREATE TABLE IF NOT EXISTS categories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    parent_id BIGINT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_active BOOLEAN,
    slug VARCHAR(255),
    icon VARCHAR(255),
    display_order INT
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    category_id BIGINT,
    price DECIMAL(10,2),
    currency VARCHAR(10),
    stock_quantity INT,
    weight DOUBLE,
    created_at TIMESTAMP,
    is_active BOOLEAN
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    total_amount DECIMAL(10,2),
    currency VARCHAR(10),
    status VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    shipping_address_id BIGINT,
    payment_method VARCHAR(50),
    notes TEXT
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT,
    product_id BIGINT,
    quantity INT,
    unit_price DECIMAL(10,2),
    total_price DECIMAL(10,2),
    discount_amount DECIMAL(10,2),
    tax_amount DECIMAL(10,2),
    created_at TIMESTAMP,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT,
    user_id BIGINT,
    rating INT,
    title VARCHAR(255),
    comment TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_verified BOOLEAN,
    helpful_count INT
);

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT,
    amount DECIMAL(10,2),
    currency VARCHAR(10),
    payment_method VARCHAR(50),
    status VARCHAR(50),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP,
    processed_at TIMESTAMP,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS inventory (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT,
    warehouse_location VARCHAR(255),
    quantity INT,
    reserved_quantity INT,
    last_updated TIMESTAMP,
    reorder_level INT,
    reorder_quantity INT,
    supplier_id BIGINT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS shipping (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT,
    carrier VARCHAR(100),
    tracking_number VARCHAR(255),
    status VARCHAR(50),
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    estimated_delivery TIMESTAMP,
    shipping_cost DECIMAL(10,2),
    notes TEXT
);

