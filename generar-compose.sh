#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <archivo_salida> <cantidad_clientes>"
    exit 1
fi

ARCHIVO_SALIDA="$1"
CANTIDAD_CLIENTES="$2"

echo "Nombre del archivo de salida: $ARCHIVO_SALIDA"
echo "Cantidad de clientes: $CANTIDAD_CLIENTES"

python3 generar_compose.py "$ARCHIVO_SALIDA" "$CANTIDAD_CLIENTES"
