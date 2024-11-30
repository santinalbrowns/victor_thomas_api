CREATE TABLE IF NOT EXISTS product_images(
    `product_id` bigint unsigned NOT NULL,
    `image_id` bigint unsigned NOT NULL,
    PRIMARY KEY (`product_id`,`image_id`),
    KEY `fk_product_images_image` (`image_id`),
    CONSTRAINT `fk_product_images_image` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_product_images_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
);