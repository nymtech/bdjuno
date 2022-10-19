#!/usr/bin/env bash
set -e

npm install --global hasura-cli

cd hasura/metadata
hasura metadata apply --endpoint http://hasura:8080 --skip-update-check --admin-secret HASURA_ADMIN_SECRET