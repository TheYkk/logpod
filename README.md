# Podlog operator


## Build
To build docker image you need to have a docker setup
You need to change username in tag, you can use private docker registry.

```
docker build -t theykk/podlog:v0.0.0 .

docker push theykk/podlog:v0.0.0
```

## Setup

```
helm install podlog podlog
```
