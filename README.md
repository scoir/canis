![Canis Major](/static/CanisMajor.jpg?raw=true "Canis Major")

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/scoir/canis/master/LICENSE)
![Build](https://github.com/scoir/canis/workflows/Build/badge.svg)
[![codecov](https://codecov.io/gh/scoir/canis/branch/master/graph/badge.svg?token=dXh8Imy2PO)](https://codecov.io/gh/scoir/canis)
[![Go Report Card](https://goreportcard.com/badge/github.com/scoir/canis)](https://goreportcard.com/report/github.com/scoir/canis)

# CANIS

Canis is an extensible credentialing platform providing a configurable, easy to use environment for issuing and verifying credentials that conform with decentralized 
identity standards including [W3C decentralized identifiers](https://w3c.github.io/did-core/) (DIDs), [W3C DID resolution](https://w3c-ccg.github.io/did-resolution/), and [W3C verifiable credentials](https://w3c.github.io/vc-data-model/).

## Summary

- [**Architecture**](#Architecture)
- [**Why Canis?**](#why-canis)
- [**Features**](#features)
- [**Building Canis**](#building-canis)
- [**Development**](#development)
- [**License**](#license)

## Architecture

![Architecture](/static/V2.png?raw=true "Canis Architecture")

Canis concerns itself levels one, two and three of the ToIP technical stack, it uses Aries to accomplish that although
we plan to support others, at level one Canis uses Indy Node. Your third party application should be level four.

## Why Canis?

Issuing digital credentials requires an institution or organization to have an agency/agent to represent its interest in the digital landscape.
This agent must act as a fiduciary on behalf of the organization, must hold cryptographic keys representing its delegated authority, and it must communicate
via [DIDComm Protocols](https://github.com/hyperledger/indy-hipe/pull/69).  

In addition, needing an agent, organizations need a way to instruct this agent how to issue a credential and to whom.  That requires information that is currently stored 
in legacy (in ToIP terms) systems.

Canis serves as a platform for creating, launching and empowering agents to participate in a credentialing ecosystem on any organization's behalf.  In addition,
Canis provides an easy to use RESTful API and extensible data model to allow for endorsing agents on behalf of any hierarchy of organizational structure.

## Features
1. **REST API**: Canis can be operated with its RESTful API for maximum flexibility
1. **Multiple Databases**: Aries today... who knows tomorrow
1. **Multiple DID Resolution**: DID resolution can be performed against...
1. **Multiple VC Formats**: Issue, prove and verify CL, JWT and JSON-LD credentials, even in the same issuance
1. **Multiple Ledger Support**:  Credential issuing on Indy, and then...
1. **Plugins**: We think the future lays in an extendable architecture for adding functionality to Canis and APIs, we're 
just starting with the ones we know best
1. **CLI**: Control your Canis platform from the command line
1. **Mailbox**: Message routing and storage for agents in support of remote, not-always-on devices

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

## License

```
Copyright 2016-2020 Scoir, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
