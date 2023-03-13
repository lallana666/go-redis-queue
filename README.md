heque
==========
* 使用go-redis实现的一个消息队列
* Job queue built on golang and etcd.

## Getting Started

```sh
go run cmd/heque-apiserver/apiserver.go -v 1 --logtostderr
```

```
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" cmd/heque-apiserver/apiserver.go
```

## Problems
* 1.redis jobs 设置超时时间1天，超时之后不再估值消费
* 2.rpoplpush，redbis或者服务中断，下次重连时，redis计数器数值没有变化，并且running中的job没有弹出
* 3.worker中job开始结束的日志打印，帮助排查线上问题
