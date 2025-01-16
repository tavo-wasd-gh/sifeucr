# SQL

## Datos de ejemplo

```sql
INSERT INTO usuarios (id_usuario, nombre) VALUES
('1@ucr.ac.cr', 'John Doe'),
('2@ucr.ac.cr', 'Clara Doe'),
('3@ucr.ac.cr', 'Jane Doe'),
('4@ucr.ac.cr', 'Bob Doe');

INSERT INTO cuentas (id_cuenta, nombre, presidencia, tesoreria, p_general, p1_servicios, p1_suministros, p1_bienes, p1_validez, p2_servicios, p2_suministros, p2_bienes, p2_validez, teeu, coes) VALUES
('DIR',     'Directorio FEUCR',                                   '1@ucr.ac.cr', '2@ucr.ac.cr', 291297589,      0,      0,      0, '2025-12-31 00:00:00',      0,      0,      0, '2025-12-31 00:00:00', 1, 1),
('F-AEFYM', 'Asociación de Estudiantes de Física y Meteorología', '3@ucr.ac.cr', '4@ucr.ac.cr',         0, 600000, 500000, 400000, '2025-07-01 00:00:00', 300000, 200000, 100000, '2025-12-01 00:00:00', 0, 1);

INSERT INTO suministros (emitido, id_cuenta, justif_sum, coes, geco, notas) VALUES
('2025-01-08 10:00:00',     'DIR', 'Horno Microondas Oster OGJ41010', 1, '2025-0001', 'Nota de ejemplo'),
('2025-04-08 10:00:00', 'F-AEFYM', 'Monitor ASUS ProArt 27"',         0, '2025-0002', 'Nota de ejemplo');

INSERT INTO servicios (emitido, id_cuenta, detalle, monto_bruto, monto_iva, monto_desc, justif_serv, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, justif_prov, coes, geco_sol, geco_ocs, por_ejecutar, ejecutado, pagado, notas) VALUES
('2025-01-07 10:00:00', 'DIR',     'Servicio de alimentación',         26000.00, 2.00, 0.00, 'Justificación ejemplo', 'Victor Avendaño',   '123456789', 'Montes de Oca', 'info@cleanco.com',        '123-456-7890', 'BCR', 'IBANCLEAN123', 'Contract approved', 1,  'SOL001', 'OCS001', '2025-02-01 00:00:00', '2025-02-10 00:00:00', NULL, 'Nota de ejemplo'),
('2025-05-06 14:00:00', 'F-AEFYM', 'Servicio de reparación de muebles', 9000.00, 2.00, 0.00, 'Justificación ejemplo', 'Ana María Bolaños', '987654321', 'Montes de Oca', 'support@itsolutions.com', '987-654-3210', 'BAC', 'IBAN ECH987',  'Service agreement', 0, 'SOL002', 'OCS002', '2025-03-01 00:00:00', NULL,                   NULL, 'Nota de ejemplo');

INSERT INTO bienes (emitido, id_cuenta, detalle, monto_bruto, monto_iva, monto_desc, justif_bien, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, justif_prov, coes, geco_sol, geco_oc, recibido, notas) VALUES
('2025-09-07 09:00:00', 'DIR',      'Sillas de oficina',          50000.00,  2.00, 0.00, 'New chairs for office', 'PC Componentes', '654321987', '789 Furniture Ave.', 'sales@furnitureco.com', '654-321-9870', 'Furniture Bank', 'IBANFURN654', 'Invoice #12345', 1,  'SOL003', 'OC003', '2025-01-15 00:00:00', 'Delivered successfully'),
('2025-10-08 11:30:00', 'F-AEFYM', 'Computadoras de escritorio', 150000.00, 2.00, 0.00, 'Replacement desktops',  'Intelec',        '321987654', '987  ech Blvd.',     'info@techsupply.com',   '321-987-6540', ' ech Bank',      'IBAN ECH321', 'Order #987',     0, 'SOL004', 'OC004', NULL, 'Pending delivery');

INSERT INTO ajustes (emitido, id_cuenta, partida, detalle, monto_bruto, notas) VALUES
('2025-01-05 12:00:00', 'F-AEFYM', 'servicios', 'Rebajo ausencia CSE', -15000.00, 'Nota');

INSERT INTO donaciones (emitido, id_cuenta_salida, partida_salida, id_cuenta_entrada, partida_entrada, detalle, monto_bruto, carta_coes, notas) VALUES
('2025-01-07 16:00:00', 'F-AEFYM', 'servicios', 'DIR', 'general', 'Donación de AEFYM para DIR para Semana U', 20000.00, 'COES-LE  ER-001', 'Notas');

INSERT INTO suministros_desglose (id_suministros, id_item, nombre, cantidad, monto_bruto_unidad) VALUES
(1, 'ITEM001', 'Pens',      10, 100.00),
(1, 'ITEM002', 'Notebooks', 2,  250.00),
(2, 'ITEM003', 'Monitors',  1,  150000.00),
(2, 'ITEM004', 'Keyboards', 3,  3000.00);
```
