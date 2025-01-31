-- Servicios

INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes)
VALUES ('2025-01-29 08:00:00', 'Carlos Martínez', 'Mantenimiento de servidores', '2025-01-30 14:00:00', 'Mantenimiento regular para evitar fallas.', FALSE);

INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes)
VALUES ('2025-01-29 09:00:00', 'Ana Gómez', 'Actualización de software en estaciones de trabajo', '2025-01-31 10:00:00', 'Actualización crítica para seguridad.', TRUE);

INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif)
VALUES ('2025-01-29 10:00:00', 'Luis Hernández', 'Instalación de nuevo sistema de videovigilancia', '2025-02-05 11:00:00', 'Proyecto de seguridad integral.', TRUE, 'Seguridad S.A.', '987654321', 'Av. Seguridad 456, Ciudad', 'seguridad@empresa.com', '555-9876', 'Banco XYZ', 'IBAN987654321', 'Cuenta habilitada para pagos de sistemas de seguridad.');

INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif, ocs_firma, ocs_firma_vive)
VALUES ('2025-01-29 11:00:00', 'Javier Rodríguez', 'Reemplazo de equipos de red obsoletos', '2025-02-07 15:00:00', 'Renovación de infraestructura de red.', TRUE, 'Redes Globales S.A.', '192837465', 'Calle Redes 789, Ciudad', 'contacto@redesglobales.com', '555-2468', 'Banco PQR', 'IBAN246813579', 'Cuenta destinada para compra de equipos de red.', 'firma_ocs_2.png', 'firma_vive_2.png');

INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif, ocs_firma, ocs_firma_vive, ejecutado, pagado, notas)
VALUES ('2025-01-29 12:00:00', 'María López', 'Desarrollo de una nueva plataforma web', '2025-02-10 16:00:00', 'Desarrollo de plataforma personalizada para el cliente.', TRUE, 'DesarrolloWeb S.A.', '102938475', 'Calle Web 101, Ciudad', 'soporte@desarrolloweb.com', '555-7531', 'Banco ABC', 'IBAN135792468', 'Pago recibido por servicios de desarrollo.', 'firma_ocs_3.png', 'firma_vive_3.png', '2025-02-08 17:00:00', '2025-02-10 18:00:00', 'Plataforma entregada y funcionando correctamente.');
