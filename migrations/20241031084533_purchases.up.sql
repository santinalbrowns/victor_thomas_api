CREATE TABLE IF NOT EXISTS purchases(
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    product_id bigint unsigned NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    quantity int NOT NULL,
    order_price FLOAT NOT NULL,
    selling_price FLOAT NOT NULL,
    store_id  bigint unsigned,
    user_id bigint unsigned,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY(`id`),
    FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`) ON DELETE SET NULL,
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
);