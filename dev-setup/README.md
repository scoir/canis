## Building Canis
Canis requires at least Go 1.14 to compile and docker >= 19.0 to create the containers required to launch Canis using either the
docker or kubernetes environment.  

It is theoretically possible to run the components of Canis (steward, agent, router, mailbox) locally by hand but that is not
a recommended approach and bypasses the power of the Canis integrated architecture.

Compiling the binaries is as simple as:

```
% make
```

After building Canis it is a good idea to run the tests:

```
% make test
```

The [Getting Started Guide](/docs/GettingStarted.md) assumes the docker execution environment.  To build
the canis docker container required for the docker execution environment, ensure docker is installed and run the following:

```
% make canis-docker
```

