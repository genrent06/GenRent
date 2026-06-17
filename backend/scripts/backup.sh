#!/usr/bin/env bash
# GenRent PostgreSQL backup script
# Usage: ./backup.sh
# Add to crontab: 0 2 * * * /path/to/backup.sh >> /var/log/genrent-backup.log 2>&1

set -euo pipefail

DB_NAME="${POSTGRES_DB:-genrent}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/genrent}"
RETAIN_DAYS="${RETAIN_DAYS:-7}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/genrent_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting backup: $BACKUP_FILE"

PGPASSWORD="${POSTGRES_PASSWORD:-}" pg_dump \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  "$DB_NAME" | gzip > "$BACKUP_FILE"

echo "[$(date)] Backup complete: $(du -sh "$BACKUP_FILE" | cut -f1)"

# Remove backups older than RETAIN_DAYS
find "$BACKUP_DIR" -name "genrent_*.sql.gz" -mtime +"$RETAIN_DAYS" -delete
echo "[$(date)] Cleaned up backups older than ${RETAIN_DAYS} days"
