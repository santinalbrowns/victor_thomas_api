CREATE TABLE IF NOT EXISTS in_store_order_details (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    order_id bigint unsigned NOT NULL,
    cashier_id bigint unsigned,
    store_id bigint unsigned NOT NULL,
    PRIMARY KEY(`id`),
    FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`cashier_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`) ON DELETE CASCADE
);