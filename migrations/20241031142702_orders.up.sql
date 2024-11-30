CREATE TABLE IF NOT EXISTS orders (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    number VARCHAR(255) UNIQUE NOT NULL,
    channel ENUM('online', 'in-store') NOT NULL,
    status ENUM('pending', 'completed', 'canceled') NOT NULL,
    total FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY(`id`)
);