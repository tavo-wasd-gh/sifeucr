-- name: GetAllBudgetEntries :many
SELECT * FROM budget_entries;

-- name: AllUsers :many
SELECT * FROM users;

-- name: UserByID :one
SELECT * FROM users
WHERE user_id = ? LIMIT 1;

-- name: UserIDByUserEmail :one
SELECT user_id FROM users
WHERE user_email = ? LIMIT 1;

-- name: ActiveUserIDByUserEmail :one
SELECT user_id FROM active_users
WHERE user_email = ? LIMIT 1;

-- name: AllAccounts :many
SELECT * FROM accounts;

-- name: AccountByID :one
SELECT * FROM accounts
WHERE account_id = ? LIMIT 1;

-- name: AllPeriods :many
SELECT * FROM periods;

-- name: AllDistributions :many
SELECT * FROM full_distributions;

-- name: AllActiveDistributions :many
SELECT * FROM full_distributions WHERE dist_active = 1;

-- name: AccountDistributions :many
SELECT * FROM full_distributions WHERE dist_account = ?;

-- name: ActiveDistributionsByAccountID :many
SELECT * FROM full_distributions WHERE dist_account = ? AND dist_active = 1;

-- name: DistributionByID :one
SELECT * FROM full_distributions
WHERE dist_id = ?;

-- name: AllSuppliers :many
SELECT * FROM suppliers;

-- name: SupplierEmails :many
SELECT supplier_email FROM suppliers;

-- name: SupplierByName :one
SELECT * FROM suppliers
WHERE supplier_name = ?
LIMIT 1;

-- name: SupplierByCatalogGrouping :one
SELECT
  s.*
FROM catalogs AS c
JOIN suppliers AS s
  ON s.supplier_id = c.catalog_supplier
WHERE c.catalog_grouping = ?
LIMIT 1;

-- name: AllCatalogs :many
SELECT * FROM full_catalogs;

-- name: CatalogByID :one
SELECT * FROM full_catalogs
WHERE catalog_id = ?;

-- name: AllCatalogItems :many
SELECT * FROM full_catalog_items;

-- name: CatalogItemByID :one
SELECT * FROM full_catalog_items
WHERE item_id = ?;

-- name: ItemAmountByID :one
SELECT item_amount FROM items
WHERE item_id = ?;

-- name: BreakdownsByPurchaseID :many
SELECT * FROM full_purchase_breakdowns
WHERE breakdown_purchase = ?;

-- name: PermissionByID :one
SELECT * FROM permissions
WHERE permission_id = ?;

-- name: AllPermissions :many
SELECT a.*, u.*, p.*
FROM users       u
JOIN permissions p ON u.user_id    = p.permission_user
JOIN accounts    a ON a.account_id = p.permission_account;

-- name: PermissionsByUserID :many
SELECT a.*, u.*, p.*
FROM users       u
JOIN permissions p ON u.user_id    = p.permission_user
JOIN accounts    a ON a.account_id = p.permission_account
WHERE u.user_id = ?;

-- name: ActivePermissionsByUserID :many
SELECT a.*, u.*, p.*
FROM active_users       u
JOIN active_permissions p ON u.user_id    = p.permission_user
JOIN active_accounts    a ON a.account_id = p.permission_account
WHERE u.user_id = ?;

-- name: PermissionByUserIDAndAccountID :one
SELECT a.*, u.*, p.*
FROM permissions p
JOIN users       u ON u.user_id    = p.permission_user
JOIN accounts    a ON a.account_id = p.permission_account
WHERE u.user_id = ? AND a.account_id = ?;

-- name: ActivePermissionByUserIDAndAccountID :one
SELECT a.*, u.*, p.*
FROM active_permissions p
JOIN active_users       u ON u.user_id    = p.permission_user
JOIN active_accounts    a ON a.account_id = p.permission_account
WHERE u.user_id = ? AND a.account_id = ?;

-- name: RequestsByAccountID :many
SELECT * FROM requests
WHERE req_account = ?;

-- name: RequestByID :one
SELECT * FROM requests
WHERE req_id = ?;

-- name: AllPurchases :many
SELECT * FROM full_purchases;

-- name: FullPurchaseByReqID :one
SELECT * FROM full_purchases
WHERE req_id = ?;

-- name: AllPurchaseSubscriptions :many
SELECT * FROM full_purchase_subscriptions;

-- name: FullPurchaseSubscriptionsByDistID :many
SELECT * FROM full_purchase_subscriptions
WHERE subscription_dist = ?;

-- name: PurchaseSubscriptionsByRequestID :many
SELECT DISTINCT *
FROM full_purchase_subscriptions
WHERE req_id = ?;

-- name: PurchaseSubscriptionByRequestIDAndAccountID :one
SELECT DISTINCT *
FROM full_purchase_subscriptions
WHERE req_id = ?
AND account_id = ?;

-- name: NewBudgetEntry :one
INSERT INTO budget_entries (
    entry_year,
    entry_code,
    entry_object,
    entry_amount
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: NewUser :one
INSERT INTO users (
    user_email,
    user_name,
    user_active
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: ToggleUserActiveByUserID :exec
UPDATE users
SET user_active = NOT user_active
WHERE user_id = ?;

-- name: AddAccount :one
INSERT INTO accounts (
    account_abbr,
    account_name,
    account_active
) VALUES (
    ?, ?, ?
)
RETURNING *;

-- name: ToggleAccountActiveByAccountID :exec
UPDATE accounts
SET account_active = NOT account_active
WHERE account_id = ?;

-- name: AddPermission :one
INSERT INTO permissions (
    permission_user,
    permission_account,
    permission_integer,
    permission_active
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: TogglePermissionByPermissionID :exec
UPDATE permissions
SET permission_integer = ?
WHERE permission_id = ?;

-- name: AddPeriod :one
INSERT INTO periods (
    period_name,
    period_start,
    period_end,
    period_active
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: UpdatePeriod :one
UPDATE periods SET
    period_name = ?,
    period_start = ?,
    period_end = ?
WHERE period_id = ? RETURNING *;

-- name: TogglePeriodActiveByPeriodID :exec
UPDATE periods
SET period_active = NOT period_active
WHERE period_id = ?;

-- name: AddDistribution :one
INSERT INTO distributions (
    dist_period,
    dist_entry_code,
    dist_account,
    dist_amount,
    dist_active
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: ToggleDistributionActiveByDistributionID :exec
UPDATE distributions
SET dist_active = NOT dist_active
WHERE dist_id = ?;

-- name: UpdateDistribution :one
UPDATE distributions SET
    dist_amount = ?
WHERE dist_id = ? RETURNING *;

-- name: AddSupplier :one
INSERT INTO suppliers (
    supplier_id,
    supplier_name,
    supplier_email,
    supplier_phone_country_code,
    supplier_phone,
    supplier_location
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpdateSupplier :one
UPDATE suppliers SET
    supplier_name = ?,
    supplier_email = ?,
    supplier_phone_country_code = ?,
    supplier_phone = ?,
    supplier_location = ?
WHERE supplier_id = ? RETURNING *;

-- name: AddCatalog :one
INSERT INTO catalogs (
    catalog_supplier,
    catalog_grouping,
    catalog_summary,
    catalog_tags
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: AddItem :one
INSERT INTO items (
    item_catalog,
    item_number,
    item_summary,
    item_description,
    item_amount
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpdateItem :one
UPDATE items SET
    item_number = ?,
    item_summary = ?,
    item_description = ?,
    item_amount = ?
WHERE item_id = ? RETURNING *;

-- name: AddRequest :one
INSERT INTO requests (
    req_user,
    req_account,
    req_issued,
    req_descr,
    req_justif
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: AddPurchase :one
INSERT INTO purchases (
    purchase_request,
    purchase_required,
    purchase_supplier,
    purchase_currency,
    purchase_ex_rate_colones,
    purchase_gross_amount,
    purchase_discount,
    purchase_tax_rate,
    purchase_geco_sol,
    purchase_geco_ord,
    purchase_bill,
    purchase_transfer,
    purchase_status
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: AddPurchaseSubscription :one
INSERT INTO purchase_subscriptions (
    "subscription_purchase",
    "subscription_user",
    "subscription_dist",
    "subscription_issued",
    "subscription_gross_amount",
    "subscription_signature",
    "subscription_signed",
    "subscription_active"
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpdatePurchaseSubscription :one
UPDATE purchase_subscriptions
SET subscription_gross_amount = ?,
    subscription_signed = ?,
    subscription_signature = ?,
    subscription_active = ?
WHERE subscription_id = ?
RETURNING *;

-- name: AddPurchaseBreakdown :one
INSERT INTO purchase_breakdowns (
    "breakdown_purchase",
    "breakdown_item",
    "breakdown_quantity"
) VALUES (
    ?, ?, ?
) RETURNING *;

-- name: PatchRequestCommon :one
UPDATE requests SET
    req_descr = ?,
    req_justif = ?
WHERE req_id = ?
RETURNING *;

-- name: PatchPurchaseCommon :one
UPDATE purchases SET
    purchase_required = ?,
    purchase_supplier = ?,
    purchase_gross_amount = ?
WHERE purchase_id = ?
RETURNING *;

-- name: PatchPurchaseMeta :one
UPDATE purchases SET
    purchase_geco_sol = ?,
    purchase_geco_ord = ?,
    purchase_bill = ?,
    purchase_transfer = ?,
    purchase_status = ?
WHERE purchase_id = ?
RETURNING *;

-- name: AddPurchaseSub :one
INSERT INTO purchase_subscriptions (
    subscription_gross_amount,
    subscription_signature,
    subscription_signed,
    subscription_active
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: PatchPurchaseSub :one
UPDATE purchase_subscriptions SET
    subscription_purchase     = ?,
    subscription_user         = ?,
    subscription_dist         = ?,
    subscription_issued       = ?,
    subscription_gross_amount = ?,
    subscription_signature    = ?,
    subscription_signed       = ?,
    subscription_active       = ?
WHERE subscription_id = ?
RETURNING *;

-- name: ToggleSubscriptionActiveByID :exec
UPDATE purchase_subscriptions
SET subscription_active = NOT subscription_active
WHERE subscription_id = ?;
