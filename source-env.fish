#!/usr/bin/env fish
# Helper script para cargar .env con formato export en fish
# Uso: source source-env.fish

if test -f .env
    # Leer el archivo .env y exportar variables usando bash
    bash -c 'source .env && env' | while read -l line
        set -l kv (string split -m 1 = $line)
        if test (count $kv) -eq 2
            set -gx $kv[1] $kv[2]
        end
    end
    echo "Variables de entorno cargadas desde .env"
else
    echo "Error: archivo .env no encontrado"
end
