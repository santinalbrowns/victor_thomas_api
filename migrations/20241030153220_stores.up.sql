CREATE TABLE IF NOT EXISTS stores (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    status BOOLEAN NOT NULL,
    PRIMARY KEY(`id`)
);