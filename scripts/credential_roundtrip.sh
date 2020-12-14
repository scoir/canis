#!/usr/bin/env bash

PATH=$(pwd)/bin:$PATH

sirius --config config/sirius-compose.yaml init --seed "b2352b32947e188eb72871093ac6217e"

sirius --config config/sirius-compose.yaml schema create --attr name:STRING --attr city:STRING --format hlindy-zkp-v1.0 --version 0.0.5 TEST2

sirius --config config/sirius-compose.yaml agents create agent-0 --public-did --schema-name TEST2
sirius --config config/sirius-compose.yaml connections invitations get agent-0 --subject indy-wallet > issuer-invite.json
curl --data-binary "@issuer-invite.json" -H "Content-Type: application/json" http://local.scoir.com:3002/connect-to-issuer

sirius --config config/sirius-compose.yaml agents create agent-1
sirius --config config/sirius-compose.yaml connections invitations get agent-1 --subject indy-wallet > verifier-invite.json
curl --data-binary "@verifier-invite.json" -H "Content-Type: application/json" http://local.scoir.com:3002/connect-to-verifier

echo "Do you wish to issue credential? (wait for connections to complete)"
select yn in "Y" "N"; do
    case $yn in
        Y ) sirius --config config/sirius-compose.yaml credentials issue agent-0 --subject indy-wallet --schema-name TEST2 --comment "this is a test" --attr name=Phil --attr city=Durham;;
        N ) exit;;
    esac
done

echo "Do you wish to request proof verfication? (wait for credential issuance to complete)"
select yn in "Y" "N"; do
    case $yn in
        Y ) sirius --config config/sirius-compose.yaml credentials request-proof agent-1 --subject indy-wallet --schema-name TEST2 --attr name= --attr city=;;
        N ) exit;;
    esac
done

