
# Scheduling Scaler

指定時刻でスケールさせるためのCRD

勉強用で作ったためバグあると思います :innocent:

## Install

RBAC無効の場合

```sh
kubectl apply -f artifacts/crd.yaml -f artifacts/crd-deployment.yaml
```

RBAC有効の場合

```sh
kubectl apply -f artifacts/crd.yaml -f artifacts/crd-deployment-rbac.yaml
```

## Usage

```yaml
apiVersion: scaling.k8s.cstoku.me/v1alpha1
kind: SchedulingScaler
metadata:
  name: example-scheduling-scaler
spec:
  location: "Asia/Tokyo"
  schedules:
  - scheduleTime: "12:00"
    replicas: 10
  - scheduleTime: "18:00"
    replicas: 15
  - scheduleTime: "23:00"
    replicas: 5
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx
```

HorizonalPodAutoscalerと雰囲気似ています。 `scaleTargetRef` とか。

`scheduleTime` で指定した時刻以降の数を `replicas` で指定する。

`location` で `scheduleTime` の時刻のタイムゾーンを指定する。

`scaleTargetRef` で時刻スケールさせる対象のリソースを指定する。
