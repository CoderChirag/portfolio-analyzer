CREATE TABLE IF NOT EXISTS `users` (
  `id` BINARY(16) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(254) NOT NULL,
  `pan` CHAR(10) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  UNIQUE KEY `uq_users_email`(`email`),
  UNIQUE KEY `uq_users_pan`(`pan`),
  CONSTRAINT `chk_users_pan` CHECK (`pan` REGEXP '^[A-Z]{5}[0-9]{4}[A-Z]$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `data_sources` (
  `id` SMALLINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  UNIQUE KEY `uq_data_sources_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `asset_types` (
  `id` SMALLINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  UNIQUE KEY `uq_asset_types` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `instruments` (
  `id` BINARY(16) NOT NULL,
  `asset_type_id` SMALLINT UNSIGNED NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `isin` CHAR(12) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `symbol` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `exchange` VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_instruments_isin` (`isin`),
  KEY `idx_instruments_symbol` (`symbol`),
  KEY `idx_instruments_asset_type` (`asset_type_id`),
  CONSTRAINT `fk_instruments_asset_type` FOREIGN KEY (`asset_type_id`) REFERENCES `asset_types` (`id`),
  CONSTRAINT `chk_instruments_exchange` CHECK (`exchange` IS NULL OR `exchange` IN ('NSE', 'BSE')),
  CONSTRAINT `chk_instruments_isin` CHECK (`isin` IS NULL OR `isin` REGEXP '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
  -- 3 = MF (fixed seed id). Every non-MF instrument must carry an ISIN.
  CONSTRAINT `chk_instruments_isin_required` CHECK (`asset_type_id` = 3 OR ISIN IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `instrument_aliases` (
  `id` BINARY(16) NOT NULL,
  `data_source_id` SMALLINT UNSIGNED NOT NULL,
  `alias` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `normalized_alias` VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `instrument_id` BINARY(16) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_aliases_source_alias` (`data_source_id`, `alias`),
  KEY `idx_aliases_normalized` (`normalized_alias`),
  KEY `idx_aliases_instrument` (`instrument_id`),
  CONSTRAINT `fk_aliases_source` FOREIGN KEY (`data_source_id`) REFERENCES `data_sources` (`id`),
  CONSTRAINT `fk_aliases_instrument` FOREIGN KEY (`instrument_id`) REFERENCES `instruments` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `data_sources` (`id`, `code`, `name`) VALUES
  (1, 'groww', 'Groww');

INSERT INTO `asset_types` (`id`, `code`, `name`) VALUES
  (1, 'STOCK', 'Stock'),
  (2, 'ETF', 'ETF'),
  (3, 'MF', 'Mutual Fund');