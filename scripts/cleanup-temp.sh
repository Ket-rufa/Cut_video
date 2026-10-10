#!/bin/bash
STORAGE_ROOT="${STORAGE_ROOT:-./storage}"
echo "Cleaning temp files older than 1 hour..."
find "$STORAGE_ROOT/temp" -type f -mmin +60 -delete 2>/dev/null || true
echo "Cleaning source files older than 24 hours..."
find "$STORAGE_ROOT/source" -type f -mmin +1440 -delete 2>/dev/null || true
echo "Cleaning exports older than 24 hours..."
find "$STORAGE_ROOT/exports" -type f -mmin +1440 -delete 2>/dev/null || true
echo "Cleanup done."
