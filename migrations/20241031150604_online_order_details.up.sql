CREATE TABLE IF NOT EXISTS online_order_details(
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    order_id bigint unsigned NOT NULL,
    customer_id bigint unsigned,
    PRIMARY KEY(`id`),
    FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`customer_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
);