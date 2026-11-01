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

-- Reference Table for keeping record of all available data sources
CREATE TABLE IF NOT EXISTS `data_sources` (
  `id` SMALLINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  UNIQUE KEY `uq_data_sources_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Reference Table for keeping record of all available instrument types
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
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_instruments_isin` (`isin`),
  KEY `idx_instruments_asset_type` (`asset_type_id`),
  FULLTEXT KEY `ft_instruments_name` (`name`),
  CONSTRAINT `fk_instruments_asset_type` FOREIGN KEY (`asset_type_id`) REFERENCES `asset_types` (`id`),
  CONSTRAINT `chk_instruments_isin` CHECK (`isin` IS NULL OR `isin` REGEXP '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
  -- 3 = MF (fixed seed id). Every non-MF instrument must carry an ISIN.
  CONSTRAINT `chk_instruments_isin_required` CHECK (`asset_type_id` = 3 OR `isin` IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Exchange listings of exchange-traded instruments (stocks, ETFs).
-- One instrument can be listed on several exchanges under different symbols.
CREATE TABLE IF NOT EXISTS `instrument_listings` (
  `id` BINARY(16) NOT NULL,
  `instrument_id` BINARY(16) NOT NULL,
  `exchange` VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `symbol` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_listings_exchange_symbol` (`exchange`, `symbol`),
  UNIQUE KEY `uq_listings_instrument_exchange` (`instrument_id`, `exchange`),
  KEY `idx_listings_symbol` (`symbol`),
  CONSTRAINT `fk_listings_instrument` FOREIGN KEY (`instrument_id`) REFERENCES `instruments` (`id`),
  CONSTRAINT `chk_listings_exchange` CHECK (`exchange` IN ('NSE', 'BSE')),
  CONSTRAINT `chk_listings_symbol` CHECK (`symbol` REGEXP '^[A-Z0-9&_-]+$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 1:1 extension of instruments for mutual funds.
CREATE TABLE IF NOT EXISTS `mutual_fund_details` (
  `instrument_id` BINARY(16) NOT NULL,
  `amc` VARCHAR(128) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`instrument_id`),
  KEY `idx_mf_details_amc` (`amc`),
  CONSTRAINT `fk_mf_details_instrument` FOREIGN KEY (`instrument_id`) REFERENCES `instruments` (`id`)
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

-- Reference Table for keeping record of all transaction types.
-- direction: +1 adds units to a holding, -1 removes them.
CREATE TABLE IF NOT EXISTS `transaction_types` (
  `id` SMALLINT UNSIGNED NOT NULL,
  `code` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `direction` TINYINT NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_transaction_types_code` (`code`),
  CONSTRAINT `chk_transaction_types_direction` CHECK (`direction` IN (-1, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Core ledger: what every asset class has. Asset-specific data lives in extension tables.
CREATE TABLE IF NOT EXISTS `transactions` (
  `id` BINARY(16) NOT NULL,
  `user_id` BINARY(16) NOT NULL,
  `data_source_id` SMALLINT UNSIGNED NOT NULL,
  `instrument_id` BINARY(16) NOT NULL,
  `transaction_type_id` SMALLINT UNSIGNED NOT NULL,
  -- Indian calendar date of the trade; drives FIFO order and FY boundaries.
  `trade_date` DATE NOT NULL,
  -- UTC instant, when the source reports a time.
  `executed_at` DATETIME(3) NULL,
  -- Always positive; the sign comes from transaction_types.direction.
  `quantity` DECIMAL(20,6) NOT NULL,
  -- Gross trade value before charges, as reported. Authoritative for cost basis.
  `amount` DECIMAL(20,4) NOT NULL,
  -- Per-unit price, only when the source reports it; never derived from amount / quantity.
  `price` DECIMAL(20,6) NULL,
  `currency` CHAR(3) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'INR',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_txn_user_instrument_date` (`user_id`, `instrument_id`, `trade_date`, `executed_at`),
  KEY `idx_txn_user_source_date` (`user_id`, `data_source_id`, `trade_date`),
  KEY `idx_txn_user_date` (`user_id`, `trade_date`),
  KEY `idx_txn_source` (`data_source_id`),
  KEY `idx_txn_instrument` (`instrument_id`),
  KEY `idx_txn_type` (`transaction_type_id`),
  CONSTRAINT `fk_txn_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_txn_source` FOREIGN KEY (`data_source_id`) REFERENCES `data_sources` (`id`),
  CONSTRAINT `fk_txn_instrument` FOREIGN KEY (`instrument_id`) REFERENCES `instruments` (`id`),
  CONSTRAINT `fk_txn_type` FOREIGN KEY (`transaction_type_id`) REFERENCES `transaction_types` (`id`),
  CONSTRAINT `chk_txn_quantity` CHECK (`quantity` > 0),
  CONSTRAINT `chk_txn_amount` CHECK (`amount` > 0),
  CONSTRAINT `chk_txn_price` CHECK (`price` IS NULL OR `price` > 0),
  CONSTRAINT `chk_txn_currency` CHECK (`currency` REGEXP '^[A-Z]{3}$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 1:1 extension of transactions for exchange-traded instruments (stocks, ETFs).
CREATE TABLE IF NOT EXISTS `stock_trade_details` (
  `transaction_id` BINARY(16) NOT NULL,
  -- Copy of transactions.trade_date: exchange order ids restart every trading day,
  -- so the dedup key needs the date. Written once by the importer, never updated.
  `trade_date` DATE NOT NULL,
  `exchange` VARCHAR(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `exchange_order_id` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`transaction_id`),
  UNIQUE KEY `uq_stock_trades_order` (`exchange`, `trade_date`, `exchange_order_id`),
  CONSTRAINT `fk_stock_trades_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`id`),
  CONSTRAINT `chk_stock_trades_exchange` CHECK (`exchange` IN ('NSE', 'BSE')),
  CONSTRAINT `chk_stock_trades_order_id` CHECK (`exchange_order_id` REGEXP '^[A-Za-z0-9]+$')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `data_sources` (`id`, `code`, `name`) VALUES
  (1, 'groww', 'Groww');

INSERT INTO `asset_types` (`id`, `code`, `name`) VALUES
  (1, 'STOCK', 'Stock'),
  (2, 'ETF', 'ETF'),
  (3, 'MF', 'Mutual Fund');

INSERT INTO `transaction_types` (`id`, `code`, `name`, `direction`) VALUES
  (1, 'BUY', 'Buy', 1),
  (2, 'SELL', 'Sell', -1);
