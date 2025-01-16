CREATE TABLE cuentas (
  id_cuenta varchar(500) NOT NULL PRIMARY KEY,
  nombre varchar(500) NOT NULL,
  presidencia varchar(500) NOT NULL,
  tesoreria varchar(500) NOT NULL,
  p_general decimal,
  p1_servicios decimal,
  p1_suministros decimal,
  p1_bienes decimal,
  p1_validez datetime NOT NULL,
  p2_servicios decimal,
  p2_suministros decimal,
  p2_bienes decimal,
  p2_validez datetime NOT NULL,
  teeu boolean NOT NULL,
  coes boolean NOT NULL,
  FOREIGN KEY (presidencia) REFERENCES usuarios (id_usuario),
  FOREIGN KEY (tesoreria) REFERENCES usuarios (id_usuario)
);

CREATE TABLE suministros (
  id_suministros integer NOT NULL PRIMARY KEY,
  emitido datetime NOT NULL,
  id_cuenta varchar(500) NOT NULL,
  justif_sum varchar(500) NOT NULL,
  coes boolean,
  geco varchar(500),
  notas varchar(500),
  FOREIGN KEY (id_cuenta) REFERENCES cuentas (id_cuenta)
);

CREATE TABLE servicios (
  id_servicios integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  emitido datetime NOT NULL,
  id_cuenta varchar(500) NOT NULL,
  detalle varchar(500) NOT NULL,
  monto_bruto decimal NOT NULL,
  monto_iva decimal NOT NULL,
  monto_desc decimal,
  justif_serv varchar(500) NOT NULL,
  prov_nom varchar(500) NOT NULL,
  prov_ced varchar(500) NOT NULL,
  prov_direc varchar(500) NOT NULL,
  prov_email varchar(500) NOT NULL,
  prov_tel varchar(500) NOT NULL,
  prov_banco varchar(500) NOT NULL,
  prov_iban varchar(500) NOT NULL,
  justif_prov varchar(500) NOT NULL,
  coes boolean,
  geco_sol varchar(500) UNIQUE,
  geco_ocs varchar(500) UNIQUE,
  por_ejecutar datetime NOT NULL,
  ejecutado datetime,
  pagado datetime,
  notas varchar(500),
  FOREIGN KEY (id_cuenta) REFERENCES cuentas (id_cuenta)
);

CREATE TABLE bienes (
  id_bienes integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  emitido datetime NOT NULL,
  id_cuenta varchar(500) NOT NULL,
  detalle varchar(500) NOT NULL,
  monto_bruto decimal NOT NULL,
  monto_iva decimal NOT NULL,
  monto_desc decimal,
  justif_bien varchar(500) NOT NULL,
  prov_nom varchar(500) NOT NULL,
  prov_ced varchar(500) NOT NULL,
  prov_direc varchar(500) NOT NULL,
  prov_email varchar(500) NOT NULL,
  prov_tel varchar(500) NOT NULL,
  prov_banco varchar(500) NOT NULL,
  prov_iban varchar(500) NOT NULL,
  justif_prov varchar(500) NOT NULL,
  coes boolean,
  geco_sol varchar(500) UNIQUE,
  geco_oc varchar(500) UNIQUE,
  recibido datetime,
  notas varchar(500),
  FOREIGN KEY (id_cuenta) REFERENCES cuentas (id_cuenta)
);

CREATE TABLE ajustes (
  id_ajustes integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  emitido datetime NOT NULL,
  id_cuenta varchar(500) NOT NULL,
  partida varchar(500) NOT NULL,
  detalle varchar(500) NOT NULL,
  monto_bruto decimal NOT NULL,
  notas varchar(500),
  FOREIGN KEY (id_cuenta) REFERENCES cuentas (id_cuenta)
);

CREATE TABLE donaciones (
  id_donacion integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  emitido datetime NOT NULL,
  id_cuenta_salida varchar(500) NOT NULL,
  partida_salida varchar(500) NOT NULL,
  id_cuenta_entrada varchar(500) NOT NULL,
  partida_entrada varchar(500) NOT NULL,
  detalle varchar(500) NOT NULL,
  monto_bruto decimal NOT NULL,
  carta_coes varchar(500) NOT NULL,
  notas varchar(500),
  FOREIGN KEY (id_cuenta_salida) REFERENCES cuentas (id_cuenta),
  FOREIGN KEY (id_cuenta_entrada) REFERENCES cuentas (id_cuenta)
);

CREATE TABLE suministros_desglose (
  id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  id_suministros integer NOT NULL,
  id_item varchar(500) NOT NULL,
  nombre varchar(500) NOT NULL,
  cantidad integer NOT NULL,
  monto_bruto_unidad decimal NOT NULL,
  FOREIGN KEY (id_suministros) REFERENCES suministros (id_suministros)
);

CREATE TABLE usuarios (
  id_usuario varchar(500) NOT NULL PRIMARY KEY,
  nombre varchar(500) NOT NULL
);
