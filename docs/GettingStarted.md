# Basics

```
. canis.sh
make canis-docker
```

Install RabbitMQ

```
sudo rabbitmqctl add_user canis canis
sudo rabbitmqctl add_vhost canis
sudo rabbitmqctl set_permissions -p "canis" "canis" ".*" ".*" ".*"
```

You start a Canis cluster using the `sirius` command line tool.  The basic command structure is:

```
% bin/sirius init --config ./config/sirius.yaml --seed "b2352b32947e188eb72871093ac6217e"
```

## Configuration

Canis has a pluggable architecture using a configuration file to tell Canis what database and execution
environment you will be using.  Other plug-ins include verifiable data registries, wallet storage engines and credential formats.

Canis requires a database and execution environment configured at a minimum.  The sample configuration located at [config/docker/canis.yaml](/config/docker/canis.yaml) assumes a MongoDB database
and the docker execution environment.  

### Datastore

The following section from the sample configuration tells Canis to use MongoDB running on localhost listening on its standard port.

``` yaml
datastore:
  database: mongo
  mongo:
     url: "mongodb://172.17.0.1:27017"
     database: "canis"
```

For Canis running in a docker container to connect to a database running on your docker host you need to find the address
of the `docker0` interface and replace `172.17.0.1` with your address.  On Linux, run the following:

```sh
% ifconfig docker0
docker0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 172.17.0.1  netmask 255.255.0.0  broadcast 172.17.255.255
        inet6 fe80::42:3eff:fed4:ee43  prefixlen 64  scopeid 0x20<link>
        ether 02:42:3e:d4:ee:43  txqueuelen 0  (Ethernet)
        RX packets 45435  bytes 9418865 (9.4 MB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 38735  bytes 67489173 (67.4 MB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

```

Use the address from `inet` in your configuration.

### Execution Environment

The sample configuration uses the docker execution environment configured in the following section:

```yaml
execution:
  runtime: docker
  docker:
    home: /tmp/canis
```

The value of `home` must already exist and be a directory to which you have write access.

### Steward

The Steward is the root of trust in a Canis cluster.  It also hosts the Admin API on both a gRPC port and an HTTP port.
Those ports are configured in the following section:

```yaml
steward:
  dbpath: /tmp/canis/steward

  wsinbound:
    host: 0.0.0.0
    port: 7777
  grpc:
    host: 0.0.0.0
    port: 7778
  grpcBridge:
    host: 0.0.0.0
    port: 7779
```


## Launch Canis

To start your cluster run:

```sh
% sirius start --config config/docker/canis.yaml
```

To check the status of the components of a canis cluster run the status command.  The output should resemble:

```sh 
% sirius status --config config/docker/canis.yaml
NAME      ID             STATUS    TIME
steward   7311b8deabe9   RUNNING   5s

Agents running: 0
```

### Admin API

Once the Steward is running, a Swagger UI is available to browse and test the Admin API.  The UI is available
on the `grpcBridge` port listed above.  Point your browser at:


[http://localhost:7779/swaggerui/](http://localhost:7779/swaggerui/)


To test agents, run the `POST /agents` endpoint to create an agent.  Take the ID used to create the agent and
then execute `POST /agents/{id}/launch` to launch an agent container.  Running the status command should indicate that
one agent is now running:

```sh 
% sirius status --config config/docker/canis.yaml
NAME      ID             STATUS    TIME
steward   7311b8deabe9   RUNNING   4m

Agents running: 1
```





