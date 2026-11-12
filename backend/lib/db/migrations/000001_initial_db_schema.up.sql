CREATE TABLE IF NOT EXISTS `users` (
  `id` BINARY(16) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(254) NOT NULL,
  `pan` CHAR(10) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `version` INT UNSIGNED NOT NULL DEFAULT 1,
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

-- Imports: What user sees
CREATE TABLE IF NOT EXISTS `imports` (
  `id` BINARY(16) NOT NULL,
  `user_id` BINARY(16) NOT NULL,
  `data_source_id` SMALLINT UNSIGNED NOT NULL,
  -- Which parser handles the file, e.g. 'groww_stock_orders'.
  `kind` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `file_name` VARCHAR(255) NOT NULL,
  `file_size` INT UNSIGNED NOT NULL,
  `file_sha256` BINARY(32) NOT NULL,
  -- queued / processing / succeeded / completed_with_errors / failed
  `status` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  -- NULL until the file has been parsed and its rows counted.
  `rows_total` INT UNSIGNED NULL,
  `rows_processed` INT UNSIGNED NOT NULL DEFAULT 0,
  `rows_imported` INT UNSIGNED NOT NULL DEFAULT 0,
  `rows_duplicate` INT UNSIGNED NOT NULL DEFAULT 0,
  `rows_failed` INT UNSIGNED NOT NULL DEFAULT 0,
  -- File-level failure only (unreadable file, retries exhausted). Row problems go to import_row_errors.
  `error_code` VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `error_msg` VARCHAR(1024) NULL,
  `started_at` DATETIME(3) NULL,
  `finished_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  UNIQUE KEY `uq_imports_user_file` (`user_id`, `file_sha256`),
  KEY `idx_imports_user_status` (`user_id`, `status`),
  KEY `idx_imports_user` (`user_id`),
  KEY `idx_imports_source` (`data_source_id`),
  CONSTRAINT `fk_imports_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_imports_source` FOREIGN KEY (`data_source_id`) REFERENCES `data_sources` (`id`),
  CONSTRAINT `chk_imports_kind` CHECK(`kind` IN ('groww_stock_orders')),
  CONSTRAINT `chk_imports_status` CHECK(`status` IN 
    ('queued', 'processing', 'succeeded', 'completed_with_errors', 'failed')),
  CONSTRAINT `chk_imports_processed` CHECK(`rows_total` IS NULL OR `rows_processed` <= `rows_total`),
  CONSTRAINT `chk_imports_finished` CHECK(`finished_at` IS NULL OR `started_at` IS NULL OR `finished_at` >= `started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Uploaded Bytes
CREATE TABLE IF NOT EXISTS `import_files` (
  `import_id` BINARY(16) NOT NULL,
  -- Normalized media type ("type/subtype", lowercase, no parameters), detected by the server.
  `content_type` VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  -- Up to 16 MB; the upload handler enforces a lower limit.
  `content` MEDIUMBLOB NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`import_id`),
  CONSTRAINT `fk_import_files_import` FOREIGN KEY (`import_id`) REFERENCES `imports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Generic Work Queue
CREATE TABLE IF NOT EXISTS `jobs` (
  `id` BINARY(16) NOT NULL,
  -- e.g. 'import.process'; a registry in code maps each type to its handler.
  `job_type` VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  -- Handler input, e.g. {"import_id": "0195..."}.
  `payload` JSON NOT NULL,
  -- Makes enqueueing idempotent, e.g. 'import.process:<import_id>'. NULL = no dedup.
  `dedupe_key` VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NULL,
  -- NULL for system jobs (price refresh, cleanup).
  `user_id` BINARY(16) NULL,
  -- queued / running / succeeded / failed / cancelled
  `status` VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `attempts` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  `max_attempts` SMALLINT UNSIGNED NOT NULL DEFAULT 5,
  -- queued: run no earlier than this. running: lease expiry; past it, another worker may take over.
  `available_at` DATETIME(3) NOT NULL,
  -- Worker identity (hostname:pid:goroutine) for debugging stuck jobs.
  `locked_by` VARCHAR(128) NULL,
  `last_error` VARCHAR(2048) NULL,
  `finished_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_jobs_dedupe_key` (`dedupe_key`),
  -- Keep the claim index small: a cleanup job deletes succeeded and cancelled jobs after about 7 days,
  -- and failed ones after about 30. The table then stays proportional to recent work, 
  -- not all-time work, which keeps claims fast regardless of history.
  KEY `idx_jobs_claim` (`status`, `available_at`),
  KEY `idx_jobs_user` (`user_id`),
  CONSTRAINT `fk_jobs_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `chk_jobs_status` CHECK(`status` IN ('queued', 'running', 'succeeded', 'failed', 'cancelled')),
  CONSTRAINT `chk_jobs_attempts` CHECK(`attempts` <= `max_attempts`),
  CONSTRAINT`chk_jobs_payload` CHECK(JSON_TYPE(`payload`) = 'OBJECT')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Why particular rows are skipped
CREATE TABLE IF NOT EXISTS `import_row_errors` (
  `id` BINARY(16) NOT NULL,
  `import_id` BINARY(16) NOT NULL,
  -- 1-based line number in the file, as a spreadsheet shows it (header = 1).
  `line_number` INT UNSIGNED NOT NULL,
  -- Column the problem is in, e.g. 'ISIN'; NULL if it concerns the whole row.
  `field` VARCHAR(64) NULL,
  -- Stable machine code for the UI, e.g. 'UNKNOWN_ISIN', 'INVALID_QUANTITY'.
  `code` VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `message` VARCHAR(1024) NOT NULL,
  -- The original row (header → value), so the user sees exactly what was rejected.
  `raw_row` JSON NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY(`id`),
  KEY `idx_import_row_errors_import_row` (`import_id`, `line_number`),
  CONSTRAINT `fk_import_row_errors_import` FOREIGN KEY (`import_id`) REFERENCES `imports` (`id`) ON DELETE CASCADE
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
  `import_id` BINARY(16) NULL,   -- NULL for manual entries
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
  KEY `idx_txn_import` (`import_id`),
  CONSTRAINT `fk_txn_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_txn_source` FOREIGN KEY (`data_source_id`) REFERENCES `data_sources` (`id`),
  CONSTRAINT `fk_txn_instrument` FOREIGN KEY (`instrument_id`) REFERENCES `instruments` (`id`),
  CONSTRAINT `fk_txn_type` FOREIGN KEY (`transaction_type_id`) REFERENCES `transaction_types` (`id`),
  CONSTRAINT `fk_txn_import` FOREIGN KEY (`import_id`) REFERENCES `imports` (`id`),
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
