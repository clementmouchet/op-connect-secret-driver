#!/usr/bin/env bash
set -euo pipefail

DRIVER="${1:-op-connect-secret-driver:latest}"

create_secret() {
  local name="$1"
  shift

  docker secret rm "$name" >/dev/null 2>&1 || true
  printf '' | docker secret create --driver "$DRIVER" "$@" "$name" - >/dev/null
  echo "created secret: $name"
}

create_secret test_username \
  --label ref="op://Test/Test Secret/username"

create_secret test_password \
  --label vault="Test" \
  --label item="Test Secret" \
  --label field="password"

create_secret test_section_password_by_attributes \
  --label vault="Test" \
  --label item="Test Secret" \
  --label section="section 1" \
  --label field="password"

create_secret test_section_password_by_ref \
  --label ref="op://Test/Test Secret/section 1/password"

create_secret test_config_yaml_by_attributes \
  --label vault="Test" \
  --label item="Test Config File" \
  --label field="config.yml"

create_secret test_config_yaml_by_ref \
  --label ref="op://Test/Test Config File/config.yml"

create_secret test_special_field_uuid_by_attributes \
  --label vault="Test" \
  --label item="buyrxgj6kdwulpomoudk3277pq" \
  --label section="qf65fxk4obzxwne6c3fithaliy" \
  --label field="2hht6727vkjyczjffo2hhflske"

create_secret test_special_field_uuid_by_ref \
  --label ref="op://Test/buyrxgj6kdwulpomoudk3277pq/qf65fxk4obzxwne6c3fithaliy/2hht6727vkjyczjffo2hhflske"
