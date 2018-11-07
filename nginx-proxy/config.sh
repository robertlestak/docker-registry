#!/bin/bash

source ../.env

function create_all () {
  awk 1 access/htpasswds/* > proxy_data/htpasswds/.htpasswd-all
}

function create_admin () {
  cp access/htpasswds/admin proxy_data/htpasswds/.htpasswd-admin
}

function create_base () {
  mkdir -p proxy_data/htpasswds proxy_data/repository-access proxy_data/conf
  create_admin
  create_all
}

function create_nginx_conf () {
  if [ ! -z "$CERT_PATH" ]
    then
      cp conf/nginx.ssl.base.conf proxy_data/conf/nginx.conf
    else
      cp conf/nginx.base.conf proxy_data/conf/nginx.conf
  fi
  sed \
    -e "s,PROXY_DOMAIN,$PROXY_DOMAIN,g" \
    -e "s,CERT_PATH,$CERT_PATH,g" \
    -e "s,CERT_KEY,$CERT_KEY,g" \
    proxy_data/conf/nginx.conf > tmp && mv tmp proxy_data/conf/nginx.conf
}

function create_access_blocks () {
  SVCS=($(ls proxy_data/htpasswds/*.htpasswd))
  for s in "${SVCS[@]}"; do
    SV=$(basename $s | awk -F'.' '{print $1}')
    s=$(basename $s)
    rm -f proxy_data/repository-access/$SV.conf
    create_access_block $SV $s >> proxy_data/repository-access/$SV.conf
  done
}

function create_access_block () {
  cat conf/repository-access-template.conf | sed \
    -e "s/{{SERVICE_REGEX}}/$1/g" \
    -e "s/{{SERVICE_HTPASSWD}}/$2/g"
}

function set_service_user () {
  echo $2 >> proxy_data/htpasswds/$1.htpasswd
}

function add_admin_to_services () {
  for a in $(ls proxy_data/htpasswds/*.htpasswd); do
    cat access/htpasswds/admin >> proxy_data/htpasswds/$(basename $a | awk -F'.' '{print $1}').htpasswd
  done
}

function create_services () {
  touch_user_services
  USERS=($(cat access/htpasswds/users))
  for up in "${USERS[@]}"; do
    THE_USER=$(echo $up | awk -F':' '{print $1}')
    THE_USERS_SERVICES=($(cat access/services/$THE_USER | xargs))
    for s in "${THE_USERS_SERVICES[@]}"; do
      set_service_user $s $up
    done
  done
  create_access_blocks
  add_admin_to_services
  cp conf/*.conf proxy_data/conf/
}

function touch_user_services () {
    USERS=($(cat access/htpasswds/users))
    for u in "${USERS[@]}"; do
      u=$(echo $u | awk -F':' '{print $1}')
      touch access/services/$u
    done
}

create_base
create_nginx_conf
touch_user_services

if [ "$1" = 'touch-services' ]; then
  exit 0
fi

create_services
