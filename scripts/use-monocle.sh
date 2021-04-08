#!/usr/bin/env bash

gcloud config set project monocle-af21d3ab
gcloud config set container/cluster phoenix
gcloud config set compute/zone us-central1-a
gcloud container clusters get-credentials phoenix
CTX=$(kubectl config current-context)
kubectl config set-context "${CTX}" --namespace=pleiades
