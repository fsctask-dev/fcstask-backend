#!/bin/sh
envsubst '${TG_SERVER_URL}' < /etc/alertmanager/alertmanager.yml.tmpl > /etc/alertmanager/alertmanager.yml
exec /bin/alertmanager "$@"
