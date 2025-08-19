CREATE TABLE IF NOT EXISTS "budget_entries" (
    "entry_id"     INTEGER PRIMARY KEY NOT NULL,
    "entry_year"   INTEGER NOT NULL,
    "entry_code"   INTEGER NOT NULL,
    "entry_object" TEXT    NOT NULL,
    "entry_amount" REAL    NOT NULL,
    UNIQUE("entry_year", "entry_code")
);

CREATE TABLE IF NOT EXISTS "users" (
    "user_id"     INTEGER PRIMARY KEY NOT NULL,
    "user_email"  TEXT    NOT NULL UNIQUE,
    "user_name"   TEXT    NOT NULL,
    "user_active" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "accounts" (
    "account_id"     INTEGER PRIMARY KEY NOT NULL,
    "account_abbr"   TEXT    NOT NULL UNIQUE,
    "account_name"   TEXT    NOT NULL UNIQUE,
    "account_active" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "permissions" (
    "permission_id"      INTEGER PRIMARY KEY NOT NULL,
    "permission_user"    INTEGER NOT NULL REFERENCES "users"("user_id"),
    "permission_account" INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "permission_integer" INTEGER NOT NULL,
    "permission_active"  BOOLEAN NOT NULL,
    UNIQUE("permission_user", "permission_account")
);

CREATE TABLE IF NOT EXISTS "periods" (
    "period_id"     INTEGER PRIMARY KEY NOT NULL,
    "period_name"   TEXT    NOT NULL UNIQUE,
    "period_start"  INTEGER NOT NULL,
    "period_end"    INTEGER NOT NULL,
    "period_active" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "distributions" (
    "dist_id"          INTEGER PRIMARY KEY NOT NULL,
    "dist_period"      INTEGER NOT NULL REFERENCES "periods"("period_id"),
    "dist_entry_code"  INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "dist_account"     INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "dist_amount"      REAL    NOT NULL,
    "dist_active"      BOOLEAN NOT NULL,
    UNIQUE("dist_period", "dist_entry_code", "dist_account")
);

CREATE TABLE IF NOT EXISTS "suppliers" (
    "supplier_id"                 INTEGER PRIMARY KEY NOT NULL,
    "supplier_name"               TEXT    NOT NULL,
    "supplier_email"              TEXT    NOT NULL,
    "supplier_phone_country_code" INTEGER NOT NULL DEFAULT '506', -- https://en.wikipedia.org/wiki/List_of_telephone_country_codes
    "supplier_phone"              INTEGER NOT NULL,
    "supplier_location"           TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS "catalogs" (
    "catalog_id"          INTEGER PRIMARY KEY NOT NULL,
    "catalog_supplier"    INTEGER NOT NULL REFERENCES "suppliers"("supplier_id"),
    "catalog_grouping"    INTEGER NOT NULL UNIQUE,
    "catalog_summary"     TEXT    NOT NULL,
    "catalog_tags"        TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS "items" (
    "item_id"          INTEGER PRIMARY KEY NOT NULL,
    "item_catalog"     INTEGER NOT NULL REFERENCES "catalogs"("catalog_id"),
    "item_number"      INTEGER NOT NULL,
    "item_summary"     TEXT    NOT NULL,
    "item_description" TEXT    NOT NULL,
    "item_amount"      REAL    NOT NULL,
    UNIQUE("item_catalog", "item_number")
);

CREATE TABLE IF NOT EXISTS "requests" (
    "req_id"      INTEGER PRIMARY KEY NOT NULL,
    "req_user"    INTEGER NOT NULL REFERENCES "users"("user_id"),
    "req_account" INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "req_issued"  INTEGER NOT NULL,
    "req_descr"   TEXT    NOT NULL,
    "req_justif"  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS "request_docs" (
    "doc_id"       INTEGER PRIMARY KEY NOT NULL,
    "doc_purchase" INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "doc_url"      TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS "global_modifications" (
    "global_mod_id"           INTEGER PRIMARY KEY NOT NULL,
    "global_mod_req"          INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "global_mod_debit_entry"  INTEGER REFERENCES "budget_entries"("entry_id"),
    "global_mod_credit_entry" INTEGER REFERENCES "budget_entries"("entry_id"),
    "global_mod_amount"       REAL    NOT NULL,
    "global_mod_letter"       TEXT
    CHECK ("global_mod_debit_entry" IS NOT NULL OR "global_mod_credit_entry" IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS "distribution_modifications" (
    "dist_mod_id"              INTEGER PRIMARY KEY NOT NULL,
    "dist_mod_request"         INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "dist_mod_debit_dist"      INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_credit_dist"     INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_amount"          REAL    NOT NULL,
    "dist_mod_letter"          TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS "purchases" (
    "purchase_id"              INTEGER PRIMARY KEY NOT NULL,
    "purchase_request"         INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "purchase_required"        INTEGER NOT NULL,
    "purchase_supplier"        INTEGER NOT NULL REFERENCES "suppliers"("supplier_id"),
    "purchase_currency"        TEXT    NOT NULL DEFAULT 'CRC', -- https://en.wikipedia.org/wiki/ISO_4217
    "purchase_ex_rate_colones" REAL    NOT NULL DEFAULT '1.00',
    "purchase_gross_amount"    REAL    NOT NULL,
    "purchase_discount"        REAL    NOT NULL DEFAULT '0.00',
    "purchase_tax_rate"        REAL    NOT NULL DEFAULT '0.02',
    "purchase_geco_sol"        TEXT    NOT NULL DEFAULT '',
    "purchase_geco_ord"        TEXT    NOT NULL DEFAULT '',
    "purchase_bill"            TEXT    NOT NULL DEFAULT '',
    "purchase_transfer"        TEXT    NOT NULL DEFAULT '',
    "purchase_status"          TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS "purchase_subscriptions" (
    "subscription_id"           INTEGER PRIMARY KEY NOT NULL,
    "subscription_purchase"     INTEGER NOT NULL REFERENCES "purchases"("purchase_id"),
    "subscription_user"         INTEGER NOT NULL REFERENCES "users"("user_id"),
    "subscription_dist"         INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "subscription_issued"       INTEGER NOT NULL,
    "subscription_gross_amount" REAL    NOT NULL,
    "subscription_signature"    TEXT    NOT NULL,
    "subscription_signed"       BOOLEAN NOT NULL,
    "subscription_active"       BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "purchase_breakdowns" (
    "breakdown_id"       INTEGER PRIMARY KEY NOT NULL,
    "breakdown_purchase" INTEGER NOT NULL REFERENCES "purchases"("purchase_id"),
    "breakdown_item"     INTEGER NOT NULL REFERENCES "items"("item_id"),
    "breakdown_quantity" REAL    NOT NULL,
    UNIQUE("breakdown_purchase", "breakdown_item")
);

CREATE VIEW IF NOT EXISTS active_accounts AS
SELECT * FROM accounts
WHERE account_active = 1;

CREATE VIEW IF NOT EXISTS active_users AS
SELECT * FROM users
WHERE user_active = 1;

CREATE VIEW IF NOT EXISTS full_distributions AS
SELECT d.*, p.*, e.*, a.*
FROM distributions  d
JOIN periods        p ON d.dist_period     = p.period_id
JOIN budget_entries e ON d.dist_entry_code = e.entry_id
JOIN accounts       a ON d.dist_account    = a.account_id;

CREATE VIEW IF NOT EXISTS active_permissions AS
SELECT * FROM permissions
WHERE permission_active = 1;

CREATE VIEW IF NOT EXISTS full_catalogs AS
SELECT c.*, s.*
FROM catalogs  c
JOIN suppliers s ON c.catalog_supplier = s.supplier_id;

CREATE VIEW IF NOT EXISTS full_catalog_items AS
SELECT i.*, c.*
FROM items         i
JOIN full_catalogs c ON i.item_catalog = c.catalog_id;

CREATE VIEW IF NOT EXISTS full_purchases AS
SELECT
  p.*,
  r.*,
  u.*,
  s.*
FROM purchases AS p
JOIN requests  AS r ON p.purchase_request = r.req_id
JOIN users     AS u ON r.req_user        = u.user_id
JOIN suppliers AS s ON p.purchase_supplier = s.supplier_id;

CREATE VIEW IF NOT EXISTS full_purchase_subscriptions AS
SELECT
    ps.*,
    p.*,
    u.*,
    d.*,
    r.*,
    be.*,
    a.*,
    per.*,
    s.*
FROM purchase_subscriptions AS ps
JOIN purchases AS p
    ON ps.subscription_purchase = p.purchase_id
JOIN users AS u
    ON ps.subscription_user = u.user_id
JOIN distributions AS d
    ON ps.subscription_dist = d.dist_id
JOIN requests AS r
    ON p.purchase_request = r.req_id
JOIN budget_entries AS be
    ON d.dist_entry_code = be.entry_id
JOIN accounts AS a
    ON d.dist_account = a.account_id
JOIN periods AS per
    ON d.dist_period = per.period_id
JOIN suppliers AS s
    ON p.purchase_supplier = s.supplier_id;

CREATE VIEW IF NOT EXISTS full_purchase_breakdowns AS
SELECT
    pb.*,
    p.*,
    i.*,
    c.*
FROM purchase_breakdowns pb
JOIN purchases p ON p.purchase_id  = pb.breakdown_purchase
JOIN items i     ON i.item_id      = pb.breakdown_item
JOIN catalogs c  ON i.item_catalog = c.catalog_id;
