-- New account
INSERT INTO accounts (id, name, teeu, coes)
VALUES ('SF', 'Secretaría de Finanzas', 1, 1);

-- New budget for account
INSERT INTO budgets (id, account, valid, services, supplies, goods)
VALUES ('SF-2025', 'SF', '2025-12-01 00:00:00', 1337095.71, 42625.00, 69297.24);

-- New user
INSERT INTO users (id, name)
VALUES ('gustavo.calvogutierrez', 'Gustavo Andrés Calvo Gutiérrez');

-- Set user permission
INSERT INTO permissions (user, account, permission_integer)
VALUES ('gustavo.calvogutierrez', 'SF', 4398046511103);
