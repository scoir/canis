#!/usr/bin/env bash

#https://unicode.org/emoji/charts/full-emoji-list.html#1f910

echo -e "\U0001F496"
echo -e "\U0001F3C3  Running additional minikube config"
echo -e "\U0001F496"
echo -e "\U0001F4C4  Updating local kubeconfig"
cat > ~/.kube/minikube-config.yaml <<EOF
apiVersion: v1
clusters:
  - cluster:
      certificate-authority: /home/$USER/.minikube/ca.crt
      server: https://$(minikube ip):8443
    name: minikube
contexts:
  - context:
      cluster: minikube
      namespace: hyades
      user: minikube
    name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
  - name: minikube
    user:
      client-certificate: /home/$USER/.minikube/client.crt
      client-key: /home/$USER/.minikube/client.key
EOF
chmod 600 ~/.kube/minikube-config.yaml
cp ~/.kube/minikube-config.yaml config/kubeconfig.yaml
chmod 600 config/kubeconfig.yaml

echo -e "\U000270D   Setting Dev IP in minikube"
export DEV_IP=172.16.1.1
minikube ssh "echo \"$DEV_IP       registry.hyades.svc.cluster.local\" | sudo tee -a  /etc/hosts" > /dev/null
minikube ssh "echo \"172.17.0.1    von-network.hyades.svc.cluster.local\" | sudo tee -a  /etc/hosts" > /dev/null

echo -e "\U0001F4DB  Creating namespace"
function create-namespace {
cat <<EOF | kubectl create -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: hyades
EOF
}
create-namespace > /dev/null

echo -e "\U0001F38F  Creating service/endpoint for Dev IP"
function create-service {
cat <<EOF | kubectl create -f -
---
kind: Service
apiVersion: v1
metadata:
  name: registry
spec:
  ports:
  - protocol: TCP
    port: 5000
    targetPort: 5000
---
kind: Endpoints
apiVersion: v1
metadata:
  name: registry
subsets:
  - addresses:
      - ip: $DEV_IP
    ports:
      - port: 5000
EOF
}
create-service > /dev/null

echo -e "\U0001F38F  Creating storage class for dynamic provisioning"
function create-storage-class {
cat <<EOF | kubectl create -f -
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
}
create-storage-class > /dev/null
kubectl create rolebinding give-it-all --clusterrole=admin --serviceaccount=hyades:default --namespace=hyades
UUID=$(cat /proc/sys/kernel/random/uuid)

CTX=$(kubectl config current-context)
kubectl config set-context "${CTX}" --namespace=hyades
