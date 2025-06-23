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

CREATE TABLE IF NOT EXISTS "distributions" (
    "dist_id"          INTEGER PRIMARY KEY NOT NULL,
    "dist_name"        TEXT    NOT NULL UNIQUE,
    "dist_entry"       INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "dist_account"     INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "dist_valid_until" INTEGER NOT NULL,
    "dist_amount"      REAL    NOT NULL,
    "dist_active"      BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "suppliers" (
    "supplier_id"                 INTEGER PRIMARY KEY NOT NULL,
    "supplier_name"               TEXT    NOT NULL,
    "supplier_email"              TEXT    NOT NULL,
    "supplier_phone_country_code" INTEGER NOT NULL DEFAULT '506', -- https://en.wikipedia.org/wiki/List_of_telephone_country_codes
    "supplier_phone"              INTEGER NOT NULL,
    "supplier_location"           TEXT
);

CREATE TABLE IF NOT EXISTS "suppliers_catalogs" (
    "catalog_id"          INTEGER PRIMARY KEY NOT NULL,
    "catalog_provider"    INTEGER NOT NULL REFERENCES "suppliers"("supplier_id"),
    "catalog_grouping"    INTEGER NOT NULL,
    "catalog_article"     INTEGER NOT NULL,
    "catalog_description" TEXT    NOT NULL,
    "catalog_amount"      REAL    NOT NULL,
    UNIQUE("catalog_grouping", "catalog_article")
);

CREATE TABLE IF NOT EXISTS "requests" (
    "req_id"      INTEGER PRIMARY KEY NOT NULL,
    "req_user"    INTEGER NOT NULL REFERENCES "users"("user_id"),
    "req_account" INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "req_issued"  INTEGER NOT NULL,
    "req_descr"   TEXT    NOT NULL,
    "req_justif"  TEXT    NOT NULL,
    "req_notes"   TEXT
);

CREATE TABLE IF NOT EXISTS "budget_transfers" (
    "transfer_id"     INTEGER PRIMARY KEY NOT NULL,
    "transfer_req"    INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "transfer_entry"  INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "transfer_amount" REAL    NOT NULL,
    "transfer_letter" TEXT
);

CREATE TABLE IF NOT EXISTS "budget_modifications" (
    "mod_id"           INTEGER PRIMARY KEY NOT NULL,
    "mod_request"      INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "mod_debit_entry"  INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "mod_credit_entry" INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "mod_amount"       REAL    NOT NULL
);

CREATE TABLE IF NOT EXISTS "distributions_modifications_types" (
    "type_id"          INTEGER PRIMARY KEY NOT NULL,
    "type_name"        TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS "distribution_modifications" (
    "dist_mod_id"              INTEGER PRIMARY KEY NOT NULL,
    "dist_mod_request"         INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "dist_mod_type"            INTEGER NOT NULL REFERENCES "distributions_modifications_types"("type_id"),
    "dist_mod_debit_dist"      INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_credit_dist"     INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_amount"          REAL    NOT NULL,
    "dist_mod_justif_approved" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "purchases" (
    "purchase_id"              INTEGER PRIMARY KEY NOT NULL,
    "purchase_request"         INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "purchase_entry"           INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "purchase_required"        INTEGER NOT NULL,
    "purchase_provider"        INTEGER REFERENCES "suppliers"("supplier_id"),
    "purchase_currency"        TEXT    DEFAULT 'CRC', -- https://en.wikipedia.org/wiki/ISO_4217
    "purchase_ex_rate_colones" REAL    DEFAULT '1.00',
    "purchase_gross_amount"    REAL,
    "purchase_discount"        REAL    DEFAULT '0.00',
    "purchase_tax_rate"        REAL    DEFAULT '0.02',
    "purchase_geco_sol"        TEXT,
    "purchase_geco_ord"        TEXT,
    "purchase_letter"          TEXT,
    "purchase_justif_approved" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "purchases_breakdown" (
    "breakdown_id"       INTEGER PRIMARY KEY NOT NULL,
    "breakdown_purchase" INTEGER NOT NULL REFERENCES "purchases"("purchase_id"),
    "breakdown_grouping" INTEGER NOT NULL REFERENCES "suppliers_catalogs"("catalog_grouping"),
    "breakdown_article"  INTEGER NOT NULL REFERENCES "suppliers_catalogs"("catalog_article"),
    "breakdown_quantity" REAL    NOT NULL,
    UNIQUE("breakdown_purchase", "breakdown_grouping", "breakdown_article")
);

CREATE VIEW IF NOT EXISTS valid_distributions AS
SELECT
    e.entry_code,
    e.entry_object,
    d.dist_id,
    d.dist_active,
    d.dist_account,
    a.account_name,
    a.account_active,
    d.dist_valid_until,
    datetime(d.dist_valid_until, 'unixepoch') AS dist_valid_until_human,
    d.dist_amount,
    d.dist_active
FROM distributions d
JOIN budget_entries e ON d.dist_entry = e.entry_id
JOIN accounts a ON d.dist_account = a.account_id
WHERE d.dist_valid_until = (
    SELECT MIN(d2.dist_valid_until)
    FROM distributions d2
    WHERE d2.dist_valid_until >= strftime('%s', 'now')
      AND d2.dist_entry = d.dist_entry
      AND d2.dist_account = d.dist_account
)
GROUP BY e.entry_code, d.dist_account;

CREATE VIEW IF NOT EXISTS active_accounts AS
SELECT
    account_id,
    account_abbr,
    account_name,
    account_active
FROM accounts
WHERE account_active = 1;

CREATE VIEW IF NOT EXISTS allowed_accounts AS
SELECT
    a.account_id,
    a.account_abbr,
    a.account_name,
    a.account_active,
    u.user_id,
    u.user_active
FROM active_accounts a
JOIN permissions p ON a.account_id = p.permission_account
JOIN users u ON p.permission_user = u.user_id
WHERE u.user_active = 1;

CREATE VIEW IF NOT EXISTS active_users AS
SELECT
    user_id,
    user_email,
    user_name,
    user_active
FROM users
WHERE user_active = 1;

CREATE VIEW IF NOT EXISTS active_permissions AS
SELECT
    permission_id,
    permission_user,
    permission_account,
    permission_integer,
    permission_active
FROM permissions
WHERE permission_active = 1;
