-- Presupuesto OPLAU
CREATE TABLE IF NOT EXISTS "budget_entries" (
    "entry_id"     INTEGER PRIMARY KEY NOT NULL,
    "entry_year"   INTEGER NOT NULL,
    "entry_code"   INTEGER NOT NULL,
    "entry_object" TEXT    NOT NULL,
    "entry_amount" REAL    NOT NULL
    -- Insert entries via yaml
);

CREATE TABLE IF NOT EXISTS "users" (
    "user_id"     INTEGER PRIMARY KEY NOT NULL,
    "user_email"  TEXT    NOT NULL,
    "user_name"   TEXT    NOT NULL,
    "user_active" BOOLEAN NOT NULL
);

-- Asocias/Órganos/Consejos
CREATE TABLE IF NOT EXISTS "accounts" (
    "account_id"     INTEGER PRIMARY KEY NOT NULL,
    "account_name"   TEXT    NOT NULL,
    "account_active" BOOLEAN NOT NULL
    -- Insert accounts via yaml
);

CREATE TABLE IF NOT EXISTS "permissions" (
    "permission_id"      INTEGER PRIMARY KEY NOT NULL,
    "permission_user"    INTEGER NOT NULL REFERENCES "users"("user_id"),
    "permission_account" INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "permission_integer" INTEGER NOT NULL,
    "permission_active"  BOOLEAN NOT NULL
);

-- Distribuciones SF & CC
CREATE TABLE IF NOT EXISTS "distributions" (
    "dist_id"          INTEGER PRIMARY KEY NOT NULL,
    "dist_entry"       INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "dist_account"     INTEGER NOT NULL REFERENCES "accounts"("account_id"),
    "dist_valid_until" INTEGER NOT NULL,
    "dist_active"      BOOLEAN NOT NULL
    -- Insert distributions via yaml
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

-- Variaciones presupuestarias desde/hacia otras unidades
CREATE TABLE IF NOT EXISTS "budget_transfers" (
    "transfer_id"     INTEGER PRIMARY KEY NOT NULL,
    "transfer_req"    INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "transfer_entry"  INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "transfer_amount" REAL    NOT NULL,
    "transfer_letter" TEXT
);

-- Variaciones en el mismo presupuesto
CREATE TABLE IF NOT EXISTS "budget_modifications" (
    "mod_id"           INTEGER PRIMARY KEY NOT NULL,
    "mod_request"      INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "mod_debit_entry"  INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "mod_credit_entry" INTEGER NOT NULL REFERENCES "budget_entries"("entry_id"),
    "mod_amount"       REAL    NOT NULL
);

-- Variaciones en el mismo presupuesto
CREATE TABLE IF NOT EXISTS "distributions_modifications_types" (
    "type_id"          INTEGER PRIMARY KEY NOT NULL,
    "type_name"        TEXT    NOT NULL
    -- Insert reasons for modifications via yaml
);

-- Modificaciones en las distribuciones
CREATE TABLE IF NOT EXISTS "distribution_modifications" (
    "dist_mod_id"              INTEGER PRIMARY KEY NOT NULL,
    "dist_mod_request"         INTEGER NOT NULL REFERENCES "requests"("req_id"),
    "dist_mod_type"            INTEGER NOT NULL REFERENCES "distributions_modifications_types"("type_id"),
    "dist_mod_debit_dist"      INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_credit_dist"     INTEGER NOT NULL REFERENCES "distributions"("dist_id"),
    "dist_mod_amount"          REAL    NOT NULL,
    "dist_mod_justif_approved" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "suppliers" (
    "supplier_id"                 INTEGER PRIMARY KEY NOT NULL,
    "supplier_name"               TEXT    NOT NULL,
    "supplier_email"              TEXT    NOT NULL,
    "supplier_phone_country_code" INTEGER NOT NULL DEAFULT '506', -- https://en.wikipedia.org/wiki/List_of_telephone_country_codes
    "supplier_phone"              INTEGER NOT NULL,
    "supplier_location"           TEXT
);

CREATE TABLE IF NOT EXISTS "suppliers_catalogs" (
    "catalog_id"          INTEGER PRIMARY KEY NOT NULL,
    "catalog_grouping"    INTEGER NOT NULL,
    "catalog_article"     INTEGER NOT NULL,
    "catalog_description" TEXT    NOT NULL,
    "catalog_amount"      REAL    NOT NULL
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
    "breakdown_quantity" REAL    NOT NULL
);
