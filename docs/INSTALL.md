# Instalar SIFEUCR

## Instalar

```sh
curl -sLO "$(curl -s https://api.github.com/repos/tavo-wasd-gh/sifeucr/releases/latest | grep 'browser_download_url' | cut -d '"' -f 4)"
tar xvf sifeucr.tar.gz
cd sifeucr
make install
```

## Configurar

```sh
vim /home/sifeucr/sifeucr/.env # véase: .example.env
cp  /home/sifeucr/sifeucr/default.db /home/sifeucr/sifeucr/db.db # o restaurar una base de datos existente
```

## Iniciar servicio

```sh
systemctl enable sifeucr # Opcional, iniciar después de boot
systemctl start sifeucr
```
