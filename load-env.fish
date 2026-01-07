# Script para cargar variables de entorno en fish shell
# Uso: source load-env.fish
# O: bash -c 'source .env' (si usas formato export)

# Cloudflare Configuration
set -gx ACCOUNT_ID "tu_account_id_aqui"
set -gx API_KEY "tu_api_key_o_token_aqui"
# API_EMAIL es opcional: solo necesario si usas API Key legacy (no tokens)
# set -gx API_EMAIL "tu@email.com"
set -gx ZONE_ID "tu_zone_id_aqui"

# Email Configuration
set -gx EMAIL "osmar@or-gm.com"
set -gx EMAIL_FROM "osmar@or-gm.com"
set -gx EMAIL_TO "osmargm1202@gmail.com"
set -gx EMAIL_PASSWORD "tu_password_aqui"
set -gx SMTP_HOST "smtp.gmail.com"
set -gx SMTP_PORT "587"

# Application Configuration
set -gx SLEEP_TIME "10"
set -gx RECORD_NAMES "orgmcr.or-gm.com,drone.or-gm.com"
set -gx DEBUG "false"
