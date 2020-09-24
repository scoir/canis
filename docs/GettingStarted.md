#Introduction

This getting started guide is for running Canis in a docker environment using `docker-compose`.  The intended user for
this approach is a developer or devops engineer testing Canis locally to see its capabilities.  It is certainly possible that
the `docker-compose` configuration could be used against a Swarm environment but that has not been thoroughly tested.  Any PRs to make
this work would be greatly appreciated.

# Basics

Assuming you built the Canis docker container locally according to [dev-setup/README.md](../dev-setup/README.md) you need to set up
a few dependencies to get Canis running.  Canis relies on [RabbitMQ](https://www.rabbitmq.com/) to buffer DIDComm messages to the various components of the system that 
handle them.  You need an installation of RabbitMQ running that you can configure to accept messages from Canis.  The default configuration
expects a user named `canis` with a password of `canis` with access to a vhost named... you guessed it...  `canis`.  Once you have a
RabbitMQ available you can run the following commands to create the necessary user, vhost and permissions.    

```
% sudo rabbitmqctl add_user canis canis
% sudo rabbitmqctl add_vhost canis
% sudo rabbitmqctl set_permissions -p "canis" "canis" ".*" ".*" ".*"
```

### Datastore and Ledgerstore

The next dependency you need to set up for the default configuration is MongoDB for the datastore (canis data model) and the 
ledgerstore (Aries wallet).  (Canis also supports CouchDB and MySQL).  The configuration for the datastore looks like the following:
 
 ```yaml
 datastore:
   database: mongo
   mongo:
     url: "mongodb://172.17.0.1:27017"
     database: "canis"
```

The configuration for the ledger store is:

```yaml
ledgerstore:
  database: mongodb
  url: "mongodb://172.17.0.1:27017"
```

You will need to update the IP address to point to your instance of mongodb.  If you are running mongodb on your docker host, 
you can get the IP address of your docker interface and replace the IP address of the mongodb urls in the configuration files.
  The follow command will list the IP address of the docker network interface on your machine.  

`
% `ip addr show docker0 | grep -Po 'inet \K[\d.]+'`
`

If mongodb is hosted elsewhere use that IP address instead.  Authentication can also be added to the mongodb URLs in the config files as needed.

## Ledger Genesis Transactions

The default configuration of canis uses an [Indy Node Network](https://github.com/hyperledger/indy-node) as the verifiable data registry
against which it verifies DIDs.  When configured to use Indy style credentials, it also uses this ledger to anchor schema and credential 
definitions for issuance and verification.
 
TODO:  How to get the genesis file of your indy network and replace the transactions current in the config files.

## Launch Canis

Once you finsh setting your configuration, you can start Canis locally using the following `docker-compose` command:

```
% docker-compose up
```

You output should begin with

```
Starting compose_canis-apiserver_1       ... done
Starting compose_canis-didcomm-issuer_1  ... done
Starting compose_canis-didcomm-doorman_1 ... done
Starting compose_canis-didcomm-lb_1      ... done
```

## Ledger Access

Once all the services have started, you will need to tell your instance of canis what DID to use when interacting with the ledger.  You
use the `sirius` command line tool to seed your Canis instance.  Use the following command with the seed used to generate a DID 
with at least ENDORSER role on your Indy network configured above.

```
% bin/sirius init --config ./config/sirius-compose.yaml --seed "<seed for DID with admin role>"
```


### Admin API

Once Canis is running, a Swagger UI is available to browse and test the Admin API.  The UI is available
on the `grpcBridge` port listed in your config.  Point your browser at:


[http://localhost:7779/swaggerui/](http://localhost:7779/swaggerui/)


To test agents, run the `POST /agents` endpoint to create an agent.  To get an invitation that can be used to establish a 
connection with your new agent, you can `POST /agents/{agent_id}/invitation/{external_id}` with `agent_id` being the ID of the agent
you created and `external_id` being the external system ID to associate to this connection.