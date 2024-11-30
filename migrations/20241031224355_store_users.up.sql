CREATE TABLE IF NOT EXISTS store_users(
    `user_id` bigint unsigned NOT NULL,
    `store_id` bigint unsigned NOT NULL,
    PRIMARY KEY (`user_id`,`store_id`),
    KEY `fk_store_users_user` (`user_id`),
    CONSTRAINT `fk_store_users_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
    CONSTRAINT `fk_store_users_store` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`)
);