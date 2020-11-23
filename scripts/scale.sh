#!/bin/bash

kubectl scale deployments/canis-apiserver --replicas=0
kubectl scale deployments/canis-didcomm-doorman --replicas=0
kubectl scale deployments/canis-didcomm-issuer --replicas=0
kubectl scale deployments/canis-didcomm-lb --replicas=0
kubectl scale deployments/canis-didcomm-verifier --replicas=0
kubectl scale deployments/canis-http-indy-resolver --replicas=0
kubectl scale deployments/canis-webhook-notifier --replicas=0
kubectl scale deployments/canis-apiserver --replicas=1
kubectl scale deployments/canis-didcomm-doorman --replicas=1
kubectl scale deployments/canis-didcomm-issuer --replicas=1
kubectl scale deployments/canis-didcomm-lb --replicas=1
kubectl scale deployments/canis-didcomm-verifier --replicas=1
kubectl scale deployments/canis-http-indy-resolver --replicas=1
kubectl scale deployments/canis-webhook-notifier --replicas=1