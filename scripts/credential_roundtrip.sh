#!/usr/bin/env bash

PATH=$(pwd)/bin:$PATH

sirius --config config/sirius-compose.yaml init --seed "b2352b32947e188eb72871093ac6217e"

sirius --config config/sirius-compose.yaml schema create --attr Name:STRING --attr City:STRING --format hlindy-zkp-v1.0 --version 0.0.4 TEST1

sirius --config config/sirius-compose.yaml agents create agent-0 --public-did --schema-name TEST1
sirius --config config/sirius-compose.yaml connections invitations get agent-0 --subject indy-wallet > issuer-invite.json
curl --data-binary "@issuer-invite.json" -H "Content-Type: application/json" http://local.scoir.com:3002/connect-to-issuer

sirius --config config/sirius-compose.yaml agents create agent-1
sirius --config config/sirius-compose.yaml connections invitations get agent-1 --subject indy-wallet > verifier-invite.json
curl --data-binary "@verifier-invite.json" -H "Content-Type: application/json" http://local.scoir.com:3002/connect-to-verifier
