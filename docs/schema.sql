CREATE TABLE usuarios (
	id varchar(80) NOT NULL PRIMARY KEY,
	nombre varchar(80) NOT NULL
);

CREATE TABLE cuentas (
	id varchar(20) NOT NULL PRIMARY KEY,
	privilegio integer NOT NULL,
	nombre varchar(120) NOT NULL,
	presidencia varchar(80),
	tesoreria varchar(80),
	pg integer UNIQUE,
	p1 integer UNIQUE,
	p2 integer UNIQUE,
	teeu boolean NOT NULL,
	coes boolean NOT NULL,
	FOREIGN KEY (presidencia) REFERENCES usuarios (id),
	FOREIGN KEY (tesoreria) REFERENCES usuarios (id),
	FOREIGN KEY (pg) REFERENCES presupuestos (id),
	FOREIGN KEY (p1) REFERENCES presupuestos (id),
	FOREIGN KEY (p2) REFERENCES presupuestos (id)
);

CREATE TABLE presupuestos (
	id integer PRIMARY KEY,
	validez datetime NOT NULL,
	total decimal NOT NULL,
	servicios decimal NOT NULL,
	suministros decimal NOT NULL,
	bienes decimal NOT NULL
);

CREATE TABLE servicios_movimientos (
	id integer PRIMARY KEY,
	movimiento integer NOT NULL,
	cuenta varchar(500) NOT NULL,
	monto decimal NOT NULL,
	firma jsonb,
	FOREIGN KEY (movimiento) REFERENCES servicios (id)
);

CREATE TABLE servicios (
	id integer PRIMARY KEY,
	-- Solicitud
	emitido datetime NOT NULL,
	emisor varchar(20) NOT NULL,
	detalle varchar(10000) NOT NULL,
	por_ejecutar datetime NOT NULL,
	justif varchar(10000) NOT NULL,
	-- COES
	coes boolean NOT NULL,
	-- OSUM
	prov_nom varchar(120),
	prov_ced varchar(20),
	prov_direc varchar(300),
	prov_email varchar(80),
	prov_tel varchar(30),
	prov_banco varchar(500),
	prov_iban varchar(500),
	prov_justif varchar(10000),
	monto_bruto decimal,
	monto_iva decimal,
	monto_desc decimal,
	geco_sol varchar(20) UNIQUE,
	geco_ocs varchar(20) UNIQUE,
	-- Final
	ejecutado datetime,
	pagado datetime,
	notas varchar(10000),
	FOREIGN KEY (emisor) REFERENCES cuentas (id)
);

CREATE TABLE suministros (
	id integer PRIMARY KEY,
	-- Solicitud
	emitido datetime NOT NULL,
	emisor varchar(20) NOT NULL,
	justif varchar(10000) NOT NULL,
	-- COES
	coes boolean,
	-- OSUM
	monto_bruto_total decimal,
	geco varchar(20),
	notas varchar(10000),
	FOREIGN KEY (emisor) REFERENCES cuentas (id)
);

CREATE TABLE suministros_desglose (
	id integer PRIMARY KEY,
	desglose integer NOT NULL,
	nombre varchar(120) NOT NULL,
	articulo varchar(40) NOT NULL,
	agrupacion varchar(40) NOT NULL,
	cantidad integer NOT NULL,
	monto_unitario decimal NOT NULL,
	FOREIGN KEY (desglose) REFERENCES suministros (id)
);

CREATE TABLE bienes_movimientos (
	id integer PRIMARY KEY,
	movimiento integer NOT NULL,
	cuenta varchar(80) NOT NULL,
	monto decimal NOT NULL,
	firma jsonb,
	FOREIGN KEY (movimiento) REFERENCES bienes (id)
);

CREATE TABLE bienes (
	id integer PRIMARY KEY,
	-- Solicitud
	emitido datetime NOT NULL,
	emisor varchar(20) NOT NULL,
	detalle varchar(10000) NOT NULL,
	por_recibir datetime NOT NULL,
	justif varchar(10000) NOT NULL,
	-- COES
	coes boolean,
	-- OSUM
	prov_nom varchar(120),
	prov_ced varchar(20),
	prov_direc varchar(300),
	prov_email varchar(80),
	prov_tel varchar(30),
	prov_banco varchar(500),
	prov_iban varchar(500),
	prov_justif varchar(10000),
	monto_bruto decimal,
	monto_iva decimal,
	monto_desc decimal,
	geco_sol varchar(20) UNIQUE,
	geco_oc varchar(20) UNIQUE,
	recibido datetime,
	notas varchar(10000),
	FOREIGN KEY (emisor) REFERENCES cuentas (id)
);

CREATE TABLE ajustes (
	id integer PRIMARY KEY,
	emitido datetime NOT NULL,
	emisor varchar(20) NOT NULL,
	cuenta varchar(20) NOT NULL,
	partida varchar(10) NOT NULL,
	detalle varchar(10000) NOT NULL,
	monto_bruto decimal NOT NULL,
	notas varchar(10000),
	FOREIGN KEY (emisor) REFERENCES cuentas (id),
	FOREIGN KEY (cuenta) REFERENCES cuentas (id)
);

CREATE TABLE donaciones (
	id integer PRIMARY KEY,
	emitido datetime NOT NULL,
	cuenta_salida varchar(20) NOT NULL,
	partida_salida varchar(10) NOT NULL,
	cuenta_entrada varchar(20) NOT NULL,
	partida_entrada varchar(10) NOT NULL,
	detalle varchar(10000) NOT NULL,
	monto_bruto decimal NOT NULL,
	carta_coes varchar(500) NOT NULL,
	notas varchar(10000),
	FOREIGN KEY (cuenta_salida) REFERENCES cuentas (id),
	FOREIGN KEY (cuenta_entrada) REFERENCES cuentas (id)
);
