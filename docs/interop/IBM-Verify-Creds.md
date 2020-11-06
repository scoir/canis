The doc for creating an account is at http://doc.ibmsecurity.verify-creds.com/develop/setup/ but just do the following ...

First go to https://agency.ibmsecurity.verify-creds.com and create/login with an IBM ID.
Enter "DigitalTrust06Builders20" for the invitation code to create your account.
Once created, click on the person icon on the top right, select "Account" to see the "Account Details" panel. Use the "Agent ID" for <ACCOUNT-ID> below and "Password" for <ACCOUNT-PASSWORD> below.

From here, you should be able to follow the cheat sheet but feel free to reach out at any time.

1. Cleanup from previous run and recreate the IBM agent with seed and default ledger on sovrin.staging

export URL=https://agency.ibmsecurity.verify-creds.com
export ACCOUNT=<ACCOUNT-ID>:<ACCOUNT-PASSWORD>
export AGENT_NAME=ibm-test-agent-for-canis
export AGENT_PASS=canispw
export AGENT=$AGENT_NAME:$AGENT_PASS
curl -u $ACCOUNT $URL/api/v1/agents/$AGENT_NAME -X DELETE
curl -u $ACCOUNT $URL/api/v1/agents -X POST -d "{\"id\": \"$AGENT_NAME\", \"password\": \"$AGENT_PASS\", \"seed\": \"1111111111111111111111IBMInterop\", \"default_ledger_name\": \"sovrin.staging\" }" -H "Content-Type: application/json"

2. Go to https://selfserve.sovrin.org an onboard the did/verkey as an endorser on the StagingNet (THIS IS ALREADY DONE FOR THE ABOVE SEED)

3. IBM agent creates an invitation

curl -u $AGENT $URL/api/v1/invitations -X POST -d '{}' -H "Content-Type: application/json"

4. Test agent accepts the invitation by scanning a QRCode

Go to https://www.the-qrcode-generator.com and generate a QRCode. Scan the QRCode with the test app.

5. Get the id of the connection to test agent, create cred schema, cred def, and send credential offer to test agent

curl -u $AGENT $URL/api/v1/connections -o conns
export CONN_ID=`cat conns | jq -r '.items[0].id'`
echo "CONN_ID: $CONN_ID"
curl -u $AGENT $URL/api/v1/credential_schemas -X POST -d '{"name": "schema1", "version": "1.0", "attrs": ["attr1", "attr2", "attr3"]}' -H "Content-Type: application/json" -o schema
export SCHEMA_ID=`cat schema | jq -r '.id'`
echo "SCHEMA_ID: $SCHEMA_ID"
curl -u $AGENT $URL/api/v1/credential_definitions -X POST -d "{\"schema_id\":\"$SCHEMA_ID\"}" -H "Content-Type: application/json" -o cred_def
export CRED_DEF_ID=`cat cred_def | jq -r '.id'`
echo "CRED_DEF_ID: $CRED_DEF_ID"
echo "Sending credential offer ..."
curl -u $AGENT $URL/api/v1/credentials -X POST -d "{\"to\": {\"id\": \"$CONN_ID\"}, \"cred_def_id\": \"$CRED_DEF_ID\", \"attributes\": {\"attr1\": \"attr1Val\", \"attr2\": \"attr2Val\", \"attr3\": \"10\"}}" -H "Content-Type: application/json"
echo "Sent credential offer"

6. Accept the offer on the test app.

7. Send the proof request to the test agent.

curl -u $AGENT $URL/api/v1/verifications -X POST -d "{\"to\": {\"id\": \"$CONN_ID\"}, \"proof_request\": {\"name\":\"FromIBMAgent\",\"version\":\"1.0\",\"requested_attributes\":{\"attr1_referent\": {\"name\":\"attr1\"}}}}" -H "Content-Type: application/json"
