CREATE TABLE IF NOT EXISTS categories(
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL,
    show_in_menu BOOLEAN NOT NULL,
    show_products BOOLEAN NOT NULL,
    image_id bigint unsigned,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`image_id`) REFERENCES `images` (`id`) ON DELETE SET NULL
);