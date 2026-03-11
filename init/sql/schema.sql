CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
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

CREATE TABLE addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    street VARCHAR(255),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    created_at TIMESTAMP,
    is_default BOOLEAN,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION
);

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
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

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    category_id BIGINT,
    price DECIMAL(10,2),
    currency VARCHAR(10),
    stock_quantity INT,
    weight DOUBLE PRECISION,
    created_at TIMESTAMP,
    is_active BOOLEAN
);

CREATE TABLE product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT,
    image_url TEXT,
    alt_text VARCHAR(255),
    is_main BOOLEAN,
    created_at TIMESTAMP,
    width INT,
    height INT,
    format VARCHAR(20),
    size_kb INT
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    address_id BIGINT,
    status VARCHAR(50),
    total_amount DECIMAL(12,2),
    currency VARCHAR(10),
    payment_method VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    shipping_cost DECIMAL(10,2)
);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    product_id BIGINT,
    product_name VARCHAR(255),
    price DECIMAL(10,2),
    quantity INT,
    total_price DECIMAL(12,2),
    created_at TIMESTAMP,
    discount DECIMAL(10,2),
    tax DECIMAL(10,2)
);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    amount DECIMAL(12,2),
    currency VARCHAR(10),
    payment_method VARCHAR(50),
    payment_status VARCHAR(50),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP,
    provider VARCHAR(100),
    fee DECIMAL(10,2)
);

CREATE TABLE reviews (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    product_id BIGINT,
    rating INT,
    title VARCHAR(255),
    comment TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_verified BOOLEAN,
    helpful_votes INT
);

CREATE TABLE inventory_logs (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT,
    change_type VARCHAR(50),
    quantity_change INT,
    old_quantity INT,
    new_quantity INT,
    reason VARCHAR(255),
    created_at TIMESTAMP,
    employee_id BIGINT,
    warehouse_id BIGINT
);