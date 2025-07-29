# SIFEUCR

## Install

### Create user and directories

``` sh
useradd -m sifeucr
mkdir -p /etc/sifeucr /var/lib/sifeucr/data

cp /path/to/config.env /etc/sifeucr/config.env
cp /path/to/init.db /var/lib/sifeucr/db.db
cp /path/to/sifeucr-bin /usr/local/bin/sifeucr

chown -R sifeucr:sifeucr /etc/sifeucr && chmod 770 /etc/sifeucr
chown -R sifeucr:sifeucr /var/lib/sifeucr && chmod 750 /var/lib/sifeucr
```

## v2.6.0 Plugins

- [ ] Verificar automáticamente estado de proveedores

## v2.5.0 Incentivos de uso

- [ ] Nombre personalizado
- [ ] Notas flotantes con colores
- [ ] Chat de soporte

## v2.4.0 FSE

- [ ] Emisión de formularios
- [ ] Revisión de formularios
- [ ] Verificación de la asignación del aplicante

## v2.3.0 Implementar MinIO

- [ ] Adjuntar pruebas prov al día en entidades públicas
- [ ] Otros archivos

## v2.2.0 Calidad de vida

- [ ] Guardar y no solicitar de una vez
- [ ] Ejemplos de solicitudes
- [ ] Proveedores y catálogos

## v2.1.0 Abstraer gestión de BD

- [ ] Panel de Control
- [ ] Manejo de usuarios
- [ ] Manejo de cuentas
- [ ] Manejo de presupuestos

## v2.0.0 Segunda iteración

- [ ] Funcionalidad mínima
- [ ] Revisar que las vistas funcionen sin errores
- [ ] Triggers para auditoría

### Reportes

- [ ] Invividual Report
- [ ] Global Report
- [ ] Request dialog

### Solicitudes

- [ ] Compras - Servicios, Suministros, Activos
- [ ] Modificaciones globales
- [ ] Modificaciones internas

### Inserciones

- [ ] Presupuesto
- [ ] Usuario
- [ ] Cuenta
- [ ] Distribución
- [ ] Proveedor
