#!/bin/sh

if [ -z "$PROXY_DOMAIN" ]
  then
    echo "Domain Required"
    exit 1
fi

if [ "$PROXY_SSL" = true ]
  then
    cp nginx.ssl.base.conf nginx.conf
  else
    cp nginx.base.conf nginx.conf
fi

sed -i -e "s/PROXY_DOMAIN/$PROXY_DOMAIN/g" nginx.conf

echo "Build Complete:" $PROXY_DOMAIN
