CREATE TABLE usuarios (
  id varchar(80) NOT NULL PRIMARY KEY,
  nombre varchar(80) NOT NULL
);

CREATE TABLE cuentas (
  id varchar(20) PRIMARY KEY,
  privilegio usuarios NOT NULL,
  nombre varchar(120) NOT NULL,
  presidencia varchar(80),
  tesoreria varchar(80),
  teeu boolean NOT NULL,
  coes boolean NOT NULL,
  FOREIGN KEY (presidencia) REFERENCES usuarios (id),
  FOREIGN KEY (tesoreria) REFERENCES usuarios (id)
);

CREATE TABLE presupuestos (
  id varchar(50) PRIMARY KEY,
  cuenta varchar(20) NOT NULL,
  validez datetime NOT NULL,
  general decimal NOT NULL,
  servicios decimal NOT NULL,
  suministros decimal NOT NULL,
  bienes decimal NOT NULL,
  FOREIGN KEY (cuenta) REFERENCES cuentas (id)
);

CREATE TABLE servicios (
  id integer PRIMARY KEY,
  -- Solicitud
  emitido datetime NOT NULL,
  emisor varchar(80) NOT NULL,
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
  -- ViVE
  ocs_firma varchar(500),
  ocs_firma_vive varchar(500),
  -- Ejecutado
  acuse_usuario varchar(80),
  acuse_fecha datetime,
  acuse varchar(10000),
  acuse_firma text,
  -- Final
  pagado datetime,
  notas varchar(10000),
  FOREIGN KEY (acuse_usuario) REFERENCES usuarios (id),
  FOREIGN KEY (emisor) REFERENCES usuarios (id)
);

CREATE TABLE servicios_movimientos (
  id integer PRIMARY KEY,
  servicio integer,
  usuario varchar(80),
  cuenta varchar(20) NOT NULL,
  presupuesto varchar(50) NOT NULL,
  monto decimal,
  firma text,
  FOREIGN KEY (servicio) REFERENCES servicios (id),
  FOREIGN KEY (usuario) REFERENCES usuarios (id),
  FOREIGN KEY (cuenta) REFERENCES cuentas (id),
  FOREIGN KEY (presupuesto) REFERENCES presupuestos (id)
);

CREATE TABLE suministros (
  id integer PRIMARY KEY,
  -- Solicitud
  emitido datetime NOT NULL,
  emisor varchar(80) NOT NULL,
  presupuesto varchar(50) NOT NULL,
  justif varchar(10000) NOT NULL,
  -- COES
  coes boolean,
  -- OSUM
  monto_bruto_total decimal,
  geco varchar(20),
  -- Recibido
  acuse_usuario varchar(80),
  acuse_fecha datetime,
  acuse varchar(10000),
  acuse_firma text,
  -- Final
  notas varchar(10000),
  FOREIGN KEY (emisor) REFERENCES usuarios (id),
  FOREIGN KEY (acuse_usuario) REFERENCES usuarios (id),
  FOREIGN KEY (presupuesto) REFERENCES presupuestos (id)
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

CREATE TABLE bienes (
  id integer PRIMARY KEY,
  -- Solicitud
  emitido datetime NOT NULL,
  emisor varchar(80) NOT NULL,
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
  -- ViVE
  oc_firma varchar(500),
  oc_firma_vive varchar(500),
  -- Recibido
  acuse_usuario varchar(80),
  acuse_fecha datetime,
  acuse varchar(10000),
  acuse_firma text,
  -- Final
  pagado datetime,
  notas varchar(10000),
  FOREIGN KEY (acuse_usuario) REFERENCES usuarios (id),
  FOREIGN KEY (emisor) REFERENCES usuarios (id)
);

CREATE TABLE bienes_movimientos (
  id integer PRIMARY KEY,
  bien integer,
  usuario varchar(80),
  cuenta varchar(20) NOT NULL,
  presupuesto varchar(50),
  monto decimal,
  firma text,
  FOREIGN KEY (bien) REFERENCES bienes (id),
  FOREIGN KEY (usuario) REFERENCES usuarios (id),
  FOREIGN KEY (cuenta) REFERENCES cuentas (id),
  FOREIGN KEY (presupuesto) REFERENCES presupuestos (id)
);

CREATE TABLE ajustes (
  id integer PRIMARY KEY,
  emitido datetime NOT NULL,
  emisor varchar(80) NOT NULL,
  cuenta varchar(20) NOT NULL,
  partida varchar(10) NOT NULL,
  detalle varchar(10000) NOT NULL,
  monto_bruto decimal NOT NULL,
  notas varchar(10000),
  FOREIGN KEY (emisor) REFERENCES usuarios (id),
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

CREATE TABLE historial (
  id integer PRIMARY KEY,
  emitido datetime NOT NULL,
  usuario varchar(80) NOT NULL,
  cuenta varchar(20) NOT NULL,
  detalle varchar(500) NOT NULL
);
