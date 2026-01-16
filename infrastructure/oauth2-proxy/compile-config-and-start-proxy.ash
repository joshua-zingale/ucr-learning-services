cat oauth2-proxy.conf.template | envsubst > oauth2-proxy.conf
oauth2-proxy --config oauth2-proxy.conf