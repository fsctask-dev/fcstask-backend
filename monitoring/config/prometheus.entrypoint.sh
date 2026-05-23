#!/bin/sh
envsubst '${ACTIVE_RULE}' < /etc/prometheus/rules.yml.tmpl > /etc/prometheus/rules.yml
exec /bin/prometheus "$@"
