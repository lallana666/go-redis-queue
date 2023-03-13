heque worker
==========
Job queue worker built on golang and redis.

## Getting Started

```sh
go run worker.go -v 1 --logtostderr
```

## How to deploy?

```
#############
# On dev box
#############
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" worker.go
scp ./worker dengqian@39.101.129.132:/denggotech/heque-worker-house-valuation/worker0

ssh dengqian@39.101.129.132

#############
# On server
#############
cd /denggotech

docker stop denggotech_heque-worker-house-valuation_1
docker rm denggotech_heque-worker-house-valuation_1
mv ./heque-worker-house-valuation/worker0 ./heque-worker-house-valuation/worker
chmod g+w ./heque-worker-house-valuation/worker

docker-compose up -d heque-worker-house-valuation
docker logs -f denggotech_heque-worker-house-valuation_1
exit
```


