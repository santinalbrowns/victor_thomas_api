CREATE TABLE IF NOT EXISTS order_items(
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    order_id bigint unsigned NOT NULL,
    product_id bigint unsigned NOT NULL,
    quantity int NOT NULL,
    price FLOAT NOT NULL,
    total FLOAT GENERATED ALWAYS AS (quantity * price) STORED,
    PRIMARY KEY(`id`),
    FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
);