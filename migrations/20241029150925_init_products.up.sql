CREATE TABLE IF NOT EXISTS products(
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    sku VARCHAR(255) NOT NULL UNIQUE,
    category_id bigint unsigned,
    status BOOLEAN NOT NULL,
    visibility BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(`id`),
    FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
);