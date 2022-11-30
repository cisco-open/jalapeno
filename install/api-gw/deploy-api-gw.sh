#!/bin/bash

KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

TEST=`which $KUBE`
if [ "$?" -eq 1 ]; then
    echo "$KUBE not found, exiting..."
    exit 1
fi

cd nginx-ingress-helm-operator
git checkout v1.0.0
cp examples/ns.yaml ../ns-secrets.yaml
cp examples/default-server-secret.yaml ..
cp examples/deployment-oss-min/*.yaml ..
# script is written assuming "kubectl"
if [ "${KUBE}" != 'kubectl' ]; then
    sed "s/kubectl/${KUBE}/" Makefile > Makefile.jalapeno
    # alias kubectl=${KUBE}
fi
# undeploy rule is missing kustomize step, which breaks it
sed -i -e 's/^undeploy: /undeploy: kustomize /' Makefile.jalapeno
make -f Makefile.jalapeno deploy IMG=nginx/nginx-ingress-operator:1.0.0

cd ..
# all this should be some kind of template instead of this janky stuff...but it works.
sed -i -e "s/nginxplus\: true/nginxplus\: false/" nginx-ingress-controller.yaml
sed -i -e "s/name\: nginxingress-sample/name\: nginx-ingress/" nginx-ingress-controller.yaml
sed -i '/nginxplus/a \    defaultBackendService: jalapeno-api-gw\/api-gw-default-backend' nginx-ingress-controller.yaml
sed -i '/metadata/a \  namespace: jalapeno-api-gw' nginx-ingress-controller.yaml
sed -i -e 's/my-nginx-ingress/jalapeno-api-gw/' ns.yaml
sed -i -e 's/my-nginx-ingress/jalapeno-api-gw/' global-configuration.yaml
sed -i -e "s/host\: localhost/host\: ${HOSTNAME}/" ingress-config.yaml
${KUBE} apply -f .
