CREATE TABLE accounts (
  id VARCHAR(20) PRIMARY KEY NOT NULL,
  name VARCHAR(120) UNIQUE NOT NULL,
  teeu BOOLEAN DEFAULT 0 NOT NULL,
  coes BOOLEAN DEFAULT 0 NOT NULL
);

CREATE TABLE budgets (
  id VARCHAR(50) PRIMARY KEY,
  account VARCHAR(20) NOT NULL,
  valid DATETIME NOT NULL,
  services DECIMAL NOT NULL,
  supplies DECIMAL NOT NULL,
  goods DECIMAL NOT NULL,
  FOREIGN KEY (account) REFERENCES accounts(id)
);

CREATE TABLE budget_lines (
  -- services, supplies, goods
  line VARCHAR(20) PRIMARY KEY NOT NULL
);
INSERT INTO budget_lines (line) VALUES
  ('services'),
  ('supplies'),
  ('goods');

CREATE TABLE users (
  email VARCHAR(80) PRIMARY KEY NOT NULL,
  name VARCHAR(80) NOT NULL,
  created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  disabled DATETIME
);

CREATE TABLE permissions (
  id INTEGER PRIMARY KEY NOT NULL,
  account VARCHAR(20) NOT NULL,
  user VARCHAR(80) NOT NULL,
  permission_integer INT NOT NULL,
  FOREIGN KEY (account) REFERENCES accounts(id) ON DELETE CASCADE,
  FOREIGN KEY (user) REFERENCES users(email) ON DELETE CASCADE,
  UNIQUE (account, user)
);

CREATE TABLE request_types (
  -- service, supply, good
  type VARCHAR(20) PRIMARY KEY NOT NULL
);
INSERT INTO request_types (type) VALUES
  ('service'),
  ('supply'),
  ('good');

CREATE TABLE requests (
  id INTEGER PRIMARY KEY NOT NULL,
  type VARCHAR(20) NOT NULL,
  -- Request
  issued DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  wanted DATETIME NOT NULL,
  issuer VARCHAR(80) NOT NULL,
  description varchar(10000) NOT NULL,
  justification VARCHAR(10000) NOT NULL,
  -- COES
  coes BOOLEAN DEFAULT 0 NOT NULL,
  correction VARCHAR(10000),
  -- OSUM
  geco_id VARCHAR(20),
  supplier_name VARCHAR(120),
  supplier_id VARCHAR(20),
  supplier_address VARCHAR(300),
  supplier_email VARCHAR(80),
  supplier_phone VARCHAR(30),
  supplier_bank VARCHAR(500),
  supplier_iban VARCHAR(500),
  supplier_justification VARCHAR(10000),
  gross_amount DECIMAL,
  tax_percentage DECIMAL,
  discount_amount DECIMAL,
  -- ViVE
  order_document VARCHAR(500),
  order_signed_vive VARCHAR(500),
  -- Executed
  recieved DATETIME,
  acknowledgement VARCHAR(10000),
  acknowledged_by VARCHAR(80),
  acknowledgement_signature TEXT,
  -- Final
  payed DATETIME,
  notes VARCHAR(10000),
  -- Conditions
  void BOOLEAN DEFAULT 0 NOT NULL,
  deleted BOOLEAN DEFAULT 0 NOT NULL,
  FOREIGN KEY (issuer) REFERENCES users(email),
  FOREIGN KEY (acknowledged_by) REFERENCES users(email)
);

CREATE TABLE movements (
  id INTEGER PRIMARY KEY NOT NULL,
  request INTEGER NOT NULL,
  type VARCHAR(20) NOT NULL,
  issuer VARCHAR(80) NOT NULL,
  account VARCHAR(20) NOT NULL,
  budget VARCHAR(50) NOT NULL,
  line VARCHAR(20) NOT NULL,
  gross_amount DECIMAL,
  signature TEXT,
  FOREIGN KEY (request, type) REFERENCES requests(id, type) ON DELETE CASCADE,
  FOREIGN KEY (issuer) REFERENCES users(email),
  FOREIGN KEY (account) REFERENCES accounts(id),
  FOREIGN KEY (budget) REFERENCES budgets(id),
  FOREIGN KEY (line) REFERENCES budget_lines(line)
);

CREATE TABLE supplies_breakdown (
  id INTEGER PRIMARY KEY NOT NULL,
  request INTEGER NOT NULL,
  group_id VARCHAR(40) NOT NULL,
  item_id VARCHAR(40) NOT NULL,
  description VARCHAR(500) NOT NULL,
  number INTEGER NOT NULL,
  gross_amount DECIMAL NOT NULL,
  FOREIGN KEY (request) REFERENCES requests(id)
);

CREATE TABLE adjustments (
  id INTEGER PRIMARY KEY NOT NULL,
  issued DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  issuer VARCHAR(80) NOT NULL,
  issuer_account VARCHAR(20) NOT NULL,
  affected_account VARCHAR(20) NOT NULL,
  affected_budget VARCHAR(50) NOT NULL,
  affected_line VARCHAR(20) NOT NULL,
  description VARCHAR(10000) NOT NULL,
  gross_amount DECIMAL NOT NULL,
  notes VARCHAR(10000),
  FOREIGN KEY (issuer) REFERENCES users(email),
  FOREIGN KEY (issuer_account) REFERENCES accounts(id),
  FOREIGN KEY (affected_account) REFERENCES accounts(id),
  FOREIGN KEY (affected_budget) REFERENCES budgets(id),
  FOREIGN KEY (affected_line) REFERENCES budget_lines(line)
);

CREATE TABLE donations (
  id INTEGER PRIMARY KEY,
  issued DATETIME NOT NULL,
  issuer VARCHAR(80) NOT NULL,
  issuer_account VARCHAR(20) NOT NULL,
  debited_account VARCHAR(20) NOT NULL,
  debited_budget VARCHAR(50) NOT NULL,
  debited_line VARCHAR(20) NOT NULL,
  credited_account VARCHAR(20) NOT NULL,
  credited_budget VARCHAR(50) NOT NULL,
  credited_line VARCHAR(20) NOT NULL,
  justification VARCHAR(10000) NOT NULL,
  gross_amount DECIMAL NOT NULL,
  coes_letter VARCHAR(500),
  notas VARCHAR(10000),
  FOREIGN KEY (issuer) REFERENCES users(email),
  FOREIGN KEY (issuer_account) REFERENCES accounts(id),
  FOREIGN KEY (debited_account) REFERENCES accounts(id),
  FOREIGN KEY (debited_budget) REFERENCES budgets(id),
  FOREIGN KEY (debited_line) REFERENCES budget_lines(line),
  FOREIGN KEY (credited_account) REFERENCES accounts(id),
  FOREIGN KEY (credited_budget) REFERENCES budgets(id),
  FOREIGN KEY (credited_line) REFERENCES budget_lines(line)
);

CREATE TABLE history (
  id INTEGER PRIMARY KEY NOT NULL,
  issued DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  issuer VARCHAR(80) NOT NULL,
  issuer_account VARCHAR(20) NOT NULL,
  description VARCHAR(10000) NOT NULL,
  FOREIGN KEY (issuer) REFERENCES users(email),
  FOREIGN KEY (issuer_account) REFERENCES accounts(id)
);
