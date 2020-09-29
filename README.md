# memcached-operator
Simple memcached operator

A simple Go-based memcached operator for GHC demo. 
For more information: https://sdk.operatorframework.io/docs/installation/install-operator-sdk/

## How to build and deploy

### Prereq:

1. operator-sdk v0.18.2
2. docker
3. kubectl v1.11.3+ 
4. Kubernetes test cluster


### Download the code

git clone git@github.com:swatishr/memcached-operator.git

cd memcached-operator


###  Run operator from your local system on a cluster, usually used during dev testing:
- Create CRD: `cat deploy/crds/cache.demo.com_memcacheds_crd.yaml | kubectl create -f -`
- Run operator from this project's root directory: `operator-sdk run local --watch-namespace=demo`
- Apply CR: `cat deploy/crds/cache.demo.com_v1alpha1_memcached_cr.yaml | kubectl -n <namespace> apply -f -`

### Build operator image:

operator-sdk build <image-repo/image-name:tag>

docker push <image-repo/image-name:tag>

### Run operator as a kubernetes deployment:
- In operator.yaml, replace the REPLACE_IMAGE with <image-repo/image-name:tag> used while building the image
- Deploy CRD: `cat deploy/crds/cache.demo.com_memcacheds_crd.yaml | kubectl create -f -`
- If you want to deploy it in a new namespace, create namespace first and create pull secret if needed and add it in `imagePullSecrets` section in service_account.yaml
- `cat deploy/service_account.yaml | kubectl -n <namespace> create -f -`
- `cat deploy/role.yaml | kubectl -n <namespace> create -f -`
- `cat deploy/role_binding.yaml | kubectl -n <namespace> create -f -`
- `cat deploy/operator.yaml | kubectl -n <namespace> create -f -`
- Verify that operator deployment is in running state: `kubectl -n <namespace> get deployment |grep memcached-operator`
- Apply CR: `cat deploy/crds/cache.demo.com_v1alpha1_memcached_cr.yaml | kubectl -n <namespace> apply -f -`
- Verify that the memcached pods are in running state: `kubectl -n <namespace> get pods |grep example-memcached`
