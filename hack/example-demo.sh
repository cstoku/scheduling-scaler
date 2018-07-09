#!/bin/sh

FSTSCALE=`date "+%H:%M"`
SNDSCALE=`date --date "1 minutes" "+%H:%M"`
TRDSCALE=`date --date "2 minutes" "+%H:%M"`

MANIFEST=`sed -e "s/12:00/$FSTSCALE/" \
    -e "s/18:00/$SNDSCALE/" \
    -e "s/23:00/$TRDSCALE/" \
    -e "s/replicas: 10/replicas: 2/" \
    artifacts/example-schedulingscaler.yaml`
echo -e "\e[32;1mManifest File\e[m:"
echo "$MANIFEST"

echo -e "\n\n\e[33;1mkubectl create\e[m:"
kubectl apply -f artifacts/example-deployment.yaml
echo "$MANIFEST" | \
kubectl apply -f - &&

echo -e "\n\n\e[36;1mkubectl get\e[m:"
kubectl get ss/example-scheduling-scaler
