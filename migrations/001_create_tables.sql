CREATE TABLE orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(50),
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(100),
    shardkey VARCHAR(50),
    sm_id INTEGER,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(50)
);

CREATE TABLE deliveries (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    zip VARCHAR(50),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

CREATE TABLE payments (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(100),
    amount INTEGER,
    payment_dt BIGINT,
    bank VARCHAR(100),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number VARCHAR(255),
    price INTEGER,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INTEGER,
    size VARCHAR(50),
    total_price INTEGER,
    nm_id BIGINT,
    brand VARCHAR(255),
    status INTEGER
);