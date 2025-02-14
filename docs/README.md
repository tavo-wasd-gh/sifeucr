# SIFEUCR

## v0.3.0 Abstraer gestión de bd

- [ ] Manejo de usuarios
- [ ] Manejo de cuentas
- [ ] Manejo de presupuestos

## v0.2.0 Implementar MinIO

- [ ] Pistas de auditoría
- [ ] Adjuntar pruebas prov al día en entidades públicas
- [ ] Adjuntar OCS firmada
- [ ] Adjuntar OCS firmada ViVE

## v0.1.3 Limpiar

- [ ] Presentar errores en `app-error`, quitar `http.Error` y responder con texto
- [ ] `db.go`: (poner aqui las llamadas, idealmente tomar TODOS los datos y filtrar en Go por cada solicitud)
- [ ] `usuario.go`: `UsuarioAcreditado` debería ir en `db.go` y deberia ser privada, solamente llamarla al momento de login
- [ ] `cuenta.go`: `cuentaInit` debería estar en `db.go` y aquí solamente el struct y funciones ayuda futuras (cuenta.Registrar, cuenta.Eliminar, etc)
- [ ] `servicios.go`: Todo mal, no tuve tiempo
- [ ] `servicios.go` Todo mal, no tuve tiempo
- [ ] `bienes.go` Todo mal, no tuve tiempo
- [ ] `suministros.go` Todo mal, no tuve tiempo

## v0.1.2 Sanitizar datos

- [ ] Sanitizar datos (Mínimo de caracteres, números de solicitudes, OC y OCS de GECO, datos de entidad proveedora, etc)
- [ ] Restricciones de fechas, word-lists, extension para crear nuevas solicitudes
- [ ] Facilitar solicitud (front-end helpers, manipular usuario en login para permitir mayúsculas)

## v0.1.1 Correcciones inmediatas

- [X] Corregir errores inmediatos que puedan revelarse en el proceso de pruebas

## v0.1.0 Primera versión

- [X] Funcionalidad mínima
