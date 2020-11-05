#!/bin/bash

URLPATH=/swaggerui/#/
URL=$(minikube service -n hyades canis-apiserver --url | grep 30779)
ADDY=$(echo "$URL" | sed 's/http:\/\///g')
IFS=':' read -r -a PARTS <<< "$ADDY"
cat > ./config/sirius-minikube.yaml <<EOF
###############################################################
#
#  API Server
#
###############################################################
api:
  grpc:
    host: ${PARTS[0]}
    port: 30778

EOF

echo -e "\U0001F45F  Access the Canis Swagger UI at the following URL:"
echo
echo "${URL}${URLPATH}"
echo
echo -e "\U0001F6B2  Launch agents, create schema and issue credentials!"