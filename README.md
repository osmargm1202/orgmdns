# orgmdns

Aplicación en Go que detecta automáticamente cambios en la IP pública del servidor y actualiza los registros DNS tipo A en Cloudflare, enviando notificaciones por correo electrónico cuando se realizan cambios.

## Descripción

`orgmdns` es un servicio que se ejecuta de forma continua, verificando periódicamente la IP pública del servidor donde está desplegado. Cuando detecta que la IP ha cambiado, actualiza automáticamente los registros DNS tipo A configurados en Cloudflare y envía una notificación por correo electrónico con los detalles del cambio.

## Características

- ✅ Detección automática de IP pública usando STUN (con fallback HTTP)
- ✅ Actualización automática de registros DNS tipo A en Cloudflare
- ✅ Notificaciones por correo electrónico (Gmail SMTP)
- ✅ Logs detallados en archivo y consola
- ✅ Modo debug opcional
- ✅ Configuración mediante variables de entorno
- ✅ Dockerizado y listo para producción

## Requisitos Previos

1. **Cuenta de Cloudflare** con:
   - API Token con permisos de lectura/escritura en DNS (o API Key global)
   - `ACCOUNT_ID`: ID de tu cuenta de Cloudflare
   - `ZONE_ID`: ID de la zona DNS que contiene los registros a gestionar
   - `API_KEY`: Token o API Key de Cloudflare

2. **Cuenta de correo electrónico** con:
   - **Para Gmail**: Contraseña de aplicación configurada (no la contraseña normal)
     - Para obtener una contraseña de aplicación:
       1. Ve a tu cuenta de Google
       2. Seguridad → Verificación en 2 pasos (debe estar activada)
       3. Contraseñas de aplicaciones → Generar nueva contraseña
       4. Usa esa contraseña como `EMAIL_PASSWORD`
   - **Para otros proveedores**: Usa tu contraseña SMTP normal y configura `SMTP_HOST` y `SMTP_PORT` según tu proveedor

## Variables de Entorno

| Variable | Descripción | Requerido | Ejemplo |
|----------|-------------|-----------|---------|
| `ACCOUNT_ID` | ID de cuenta de Cloudflare | Sí | `1234567890abcdef` |
| `API_KEY` | API Token o API Key de Cloudflare | Sí | `abc123...` |
| `API_EMAIL` | Email de cuenta Cloudflare (solo si usas API Key legacy) | No | `tu@email.com` |
| `ZONE_ID` | ID de la zona DNS en Cloudflare | Sí | `abcdef1234567890` |
| `EMAIL` | Email informativo (opcional) | No | `osmar@or-gm.com` |
| `EMAIL_FROM` | Dirección que envía el correo | Sí | `osmar@or-gm.com` |
| `EMAIL_TO` | Destinatario de notificaciones | Sí | `osmargm1202@gmail.com` |
| `EMAIL_PASSWORD` | Contraseña de aplicación o contraseña SMTP | Sí | `abcd efgh ijkl mnop` |
| `SMTP_HOST` | Servidor SMTP | No | `smtp.gmail.com` (default) |
| `SMTP_PORT` | Puerto SMTP | No | `587` (default) |
| `SLEEP_TIME` | Minutos entre verificaciones | No | `10` (default: 10) |
| `RECORD_NAMES` | Registros DNS a vigilar (separados por coma) | Sí | `"orgmcr.or-gm.com,drone.or-gm.com"` |
| `DEBUG` | Activar logs de depuración | No | `true` o `false` (default: `false`) |

### Notas sobre Variables

- **RECORD_NAMES**: Lista de FQDN (nombres completos de dominio) separados por comas. Los espacios se eliminan automáticamente.
- **SLEEP_TIME**: Tiempo en **minutos** entre cada verificación. Si no se especifica, se usa 10 minutos por defecto.
- **DEBUG**: Puede activarse también con el flag `--debug` al ejecutar el binario.

## Uso en Desarrollo

### Prerrequisitos

- Go 1.23 o superior
- Variables de entorno configuradas (puedes usar un archivo `.env`)

### Instalación y Ejecución

1. **Clonar e instalar dependencias**:
```bash
go mod download
go mod tidy
```

2. **Configurar variables de entorno**:
Crea un archivo `.env` (ya está en `.gitignore`):
```bash
ACCOUNT_ID=tu_account_id
API_KEY=tu_api_key
ZONE_ID=tu_zone_id
EMAIL_FROM=tu_email@gmail.com
EMAIL_TO=destinatario@gmail.com
EMAIL_PASSWORD=tu_password
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SLEEP_TIME=10
RECORD_NAMES="orgmcr.or-gm.com,drone.or-gm.com"
DEBUG=false
```

3. **Cargar variables y ejecutar**:
```bash
export $(cat .env | xargs)
go run ./cmd/orgmdns
```

O usar el Makefile:
```bash
make run          # Ejecución normal
make run-debug    # Ejecución con debug
```

4. **Compilar binario local**:
```bash
make build
./bin/orgmdns
```

## Uso con Docker

### Build y Push

1. **Construir imagen**:
```bash
make docker-build
```

2. **Subir al registry**:
```bash
make docker-push
```

Asegúrate de estar autenticado en el registry:
```bash
docker login orgmcr.or-gm.co
```

3. **Ejecutar localmente con Docker**:
```bash
make docker-run        # Normal
make docker-run-debug  # Con debug
```

## Uso con docker-compose

El archivo `docker-compose.yml` está listo para desplegar en el servidor.

### En el Servidor

1. **Copiar archivos necesarios**:
   - `docker-compose.yml`
   - Archivo `.env` con las variables de entorno (o definirlas directamente en el compose)

2. **Levantar el servicio**:
```bash
docker compose up -d
```

3. **Ver logs**:
```bash
docker compose logs -f orgmdns
```

4. **Reiniciar servicio**:
```bash
docker compose restart orgmdns
```

5. **Detener servicio**:
```bash
docker compose down
```

### Ejemplo de `.env` para docker-compose

```bash
ACCOUNT_ID=1234567890abcdef
API_KEY=abc123...
ZONE_ID=abcdef1234567890
EMAIL=osmar@or-gm.com
EMAIL_FROM=osmar@or-gm.com
EMAIL_TO=osmargm1202@gmail.com
EMAIL_PASSWORD=abcd efgh ijkl mnop
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SLEEP_TIME=10
RECORD_NAMES="orgmcr.or-gm.com,drone.or-gm.com"
DEBUG=false
```

## Logs y Debug

### Ubicación de Logs

Los logs se guardan en el directorio `logs/`:
- **Archivo**: `logs/app.log`
- El directorio se crea automáticamente si no existe

### Niveles de Log

- **Info**: Siempre visible en consola y archivo (información general, cambios de IP, etc.)
- **Error**: Siempre visible en consola y archivo (errores de API, SMTP, etc.)
- **Debug**: Solo visible cuando `DEBUG=true` o se usa `--debug` (detalles de cada ciclo, comparaciones, etc.)

### Activar Debug

**Opción 1**: Variable de entorno
```bash
export DEBUG=true
./bin/orgmdns
```

**Opción 2**: Flag `--debug`
```bash
./bin/orgmdns --debug
```

**Opción 3**: En docker-compose
```yaml
environment:
  - DEBUG=true
# O descomentar la línea command: ["--debug"]
```

### Contenido de los Logs

Los logs incluyen:
- IP pública detectada en cada ciclo
- Registros DNS procesados
- Comparaciones de IP (solo en debug)
- Actualizaciones realizadas
- Errores de Cloudflare API
- Errores de envío de correo (no detienen la ejecución)

## Notas de Seguridad

⚠️ **IMPORTANTE**: Protege tus credenciales

- **NUNCA** commitees archivos `.env` o credenciales al repositorio
- El archivo `.env` ya está en `.gitignore`
- Usa secretos de Docker o variables de entorno del sistema en producción
- La `EMAIL_PASSWORD` debe ser una contraseña de aplicación (para Gmail) o contraseña SMTP segura
- El `API_KEY` de Cloudflare debe tener permisos mínimos necesarios (solo DNS de la zona específica)

## API de Cloudflare

La aplicación usa la API v4 de Cloudflare con soporte para dos métodos de autenticación:

- **Endpoint base**: `https://api.cloudflare.com/client/v4`

### Métodos de Autenticación

1. **API Token (Recomendado)** - Método por defecto:
   - Usa `Authorization: Bearer <API_KEY>`
   - No requiere `API_EMAIL`
   - Más seguro y recomendado por Cloudflare
   - Para crear un token: Cloudflare Dashboard → My Profile → API Tokens → Create Token

2. **API Key + Email (Legacy)** - Si tienes problemas con tokens:
   - Usa `X-Auth-Email: <API_EMAIL>` y `X-Auth-Key: <API_KEY>`
   - Requiere configurar `API_EMAIL` con tu email de Cloudflare
   - Para obtener tu API Key: Cloudflare Dashboard → My Profile → API Tokens → Global API Key

**Operaciones**:
  - `GET /zones/{zone_id}/dns_records?type=A&name={name}`: Obtener registro DNS
  - `PATCH /zones/{zone_id}/dns_records/{record_id}`: Actualizar IP del registro

## Detección de IP Pública

La aplicación usa dos métodos para obtener la IP pública:

1. **STUN** (método principal): Consulta a `stun.l.google.com:19302` por UDP
2. **HTTP Fallback**: Si STUN falla, intenta con servicios HTTP:
   - `https://api.ipify.org?format=text`
   - `https://icanhazip.com`
   - `https://ifconfig.me/ip`

## Notificaciones por Correo

Cuando se actualiza un registro DNS, se envía un correo con:

- **Asunto**: `[orgmdns] DNS actualizado: <nombre_del_registro>`
- **Contenido**:
  - Nombre del registro modificado
  - IP anterior
  - IP nueva
  - Fecha y hora del cambio

**Configuración SMTP**:
- Host: Configurable con `SMTP_HOST` (default: `smtp.gmail.com`)
- Puerto: Configurable con `SMTP_PORT` (default: `587` para STARTTLS)
- Autenticación: Usuario `EMAIL_FROM` + `EMAIL_PASSWORD`
- Soporta cualquier proveedor SMTP (Gmail, Outlook, SendGrid, etc.)

## Troubleshooting

### Error: "ACCOUNT_ID es requerido"
- Verifica que todas las variables de entorno estén configuradas correctamente

### Error: "no se pudo obtener IP pública"
- Verifica conectividad de red
- Si hay firewall, puede que STUN esté bloqueado (se usará fallback HTTP automáticamente)

### Error: "API retornó error" o "Unable to authenticate request"
- **Si usas API Token (recomendado)**:
  - Verifica que el `API_KEY` sea un token válido (no una API Key global)
  - Verifica que el token tenga permisos de lectura/escritura en DNS para la zona específica
  - No configures `API_EMAIL` si usas tokens
- **Si usas API Key legacy**:
  - Configura `API_EMAIL` con tu email de Cloudflare
  - Verifica que `API_KEY` sea tu Global API Key (no un token)
- Verifica que el `ZONE_ID` sea correcto
- Verifica que los nombres en `RECORD_NAMES` existan en Cloudflare
- Activa `DEBUG=true` para ver qué método de autenticación se está usando

### Error: "error enviando correo"
- Verifica que `EMAIL_PASSWORD` sea correcta (contraseña de aplicación para Gmail, o contraseña SMTP para otros proveedores)
- Para Gmail: verifica que la verificación en 2 pasos esté activada
- Verifica que `SMTP_HOST` y `SMTP_PORT` sean correctos para tu proveedor
- El error se registra en logs pero no detiene la aplicación

### Los logs no aparecen
- Verifica permisos de escritura en el directorio `logs/`
- En Docker, verifica que el volumen esté montado correctamente

## Makefile

Comandos disponibles:

- `make build`: Compila el binario local
- `make run`: Ejecuta la aplicación localmente
- `make run-debug`: Ejecuta con debug activado
- `make docker-build`: Construye imagen Docker
- `make docker-push`: Sube imagen al registry
- `make docker-run`: Ejecuta contenedor localmente
- `make docker-run-debug`: Ejecuta contenedor con debug
- `make clean`: Limpia binarios y logs
- `make deps`: Descarga y actualiza dependencias

## Estructura del Proyecto

```
orgmdns/
├── cmd/
│   └── orgmdns/
│       └── main.go              # Punto de entrada
├── internal/
│   ├── app/
│   │   └── runner.go            # Bucle principal
│   ├── cloudflare/
│   │   └── client.go            # Cliente API Cloudflare
│   ├── config/
│   │   └── config.go            # Configuración y variables de entorno
│   ├── ip/
│   │   └── public_ip.go         # Detección de IP pública
│   ├── logger/
│   │   └── logger.go            # Sistema de logging
│   └── notify/
│       └── email.go             # Notificaciones por correo
├── logs/                        # Logs de la aplicación (generado)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## Licencia

Este proyecto es de uso privado.

## Autor

osmargm1202
