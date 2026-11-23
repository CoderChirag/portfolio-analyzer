#!/bin/bash
set -euo pipefail

job_id=$1
if [[ $job_id =~ ^0[xX][0-9a-fA-F]{32}$ ]]; then
  job_id_sql=$job_id                       # 0x01A2...  → binary hex literal
elif [[ $job_id =~ ^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$ ]]; then
  job_id_sql="UUID_TO_BIN('$job_id')"      # 01a202cc-afd8-...  → UUID text
elif [[ $job_id =~ ^[0-9a-fA-F]{32}$ ]]; then
  job_id_sql="UNHEX('$job_id')"            # 01A202CC...  (hex without 0x)
else
  echo "Invalid job id: $job_id (expected 0x<32 hex>, UUID, or 32 hex chars)" >&2
  exit 1
fi

export MYSQL_PWD="$DB_PASSWORD"
db() { mysql -u "$DB_USER" -h "$DB_HOST" -P "$DB_PORT" -D "$DB_NAME" "$@"; }

import_id=$(db -N -s -e "
  SELECT payload FROM jobs WHERE id = $job_id_sql;
" | jq -r .importId)
[[ -n "$import_id" && "$import_id" != "null" ]] || { echo "No import found for job $job_id" >&2; exit 1; }

echo -e "Deleting import $import_id and job $job_id"
db -e "
  START TRANSACTION;
  DELETE FROM import_files WHERE import_id = UUID_TO_BIN('$import_id');
  DELETE FROM imports WHERE id = UUID_TO_BIN('$import_id');
  DELETE FROM jobs WHERE id = $job_id_sql;
  COMMIT;
"

echo DONE