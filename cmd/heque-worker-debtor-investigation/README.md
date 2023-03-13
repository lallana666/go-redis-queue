heque worker debtor investigation
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
scp ./worker zhiwei@39.101.129.132:/denggotech/heque-worker-debtor-investigation/worker0

ssh zhiwei@39.101.129.132

#############
# On server
#############
cd /denggotech

docker stop denggotech_heque-worker-debtor-investigation_1
docker rm denggotech_heque-worker-debtor-investigation_1
mv ./heque-worker-debtor-investigation/worker0 ./heque-worker-debtor-investigation/worker
chmod g+w ./heque-worker-debtor-investigation/worker

docker-compose up -d heque-worker-debtor-investigation
docker logs -f denggotech_heque-worker-debtor-investigation_1
exit
```


