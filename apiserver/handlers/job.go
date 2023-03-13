package handlers

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"go.etcd.io/etcd/clientv3"
	v3 "go.etcd.io/etcd/clientv3"
)

var ErrKeyExists = errors.New("key exists")

// EnqueueJobHandler enqueue a job into specified queue.
func EnqueueJobHandler(client *clientv3.Client, prefix string, request *restful.Request, response *restful.Response) {
	body, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		http.Error(response, "apiserver: error reading body", http.StatusBadRequest)
		return
	}
	qName := request.PathParameter("queue")
	jobID, err := newUniqueKV(client, prefix, qName, string(body))
	if err != nil {
		http.Error(response, "apiserver: error enqueue", http.StatusBadRequest)
		return
	}
	fmt.Println(jobID)
}

// DequeueJobHandler enqueue a job into specified queue.
func DequeueJobHandler(client *clientv3.Client, prefix string, request *restful.Request, response *restful.Response) {
	queue := request.PathParameter("queue")
	key := fmt.Sprintf("%s/jobs/%s", prefix, queue)
	// count keys about to be deleted
	resp, err := client.Get(context.TODO(), key, clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		http.Error(response, "apiserver: error enqueue", http.StatusBadRequest)
	}
	if resp.Kvs == nil || len(resp.Kvs) == 0 {
		http.Error(response, "apiserver: error dequeue", http.StatusNoContent)
	}
	deleteKey := string(resp.Kvs[0].Key)

	// delete the keys
	_, err = client.Delete(context.TODO(), deleteKey, clientv3.WithPrefix())
	if err != nil {
		http.Error(response, "apiserver: error dequeue", http.StatusInternalServerError)
	}
	response.Write(resp.Kvs[0].Value)
}

func newUniqueKV(kv v3.KV, prefix string, queue string, job string) (string, error) {
	for {
		jobID := strconv.FormatInt(time.Now().UnixNano(), 10)
		newKey := fmt.Sprintf("%s/jobs/%s/%s", prefix, queue, jobID)
		// TODO: 后期添加添加过期时间
		_, err := putNewKV(kv, newKey, job, v3.NoLease)
		if err != nil {
			return "", err
		}
		if err != ErrKeyExists {
			return "", nil
		}
		return jobID, nil
	}
}

// putNewKV attempts to create the given key, only succeeding if the key did
// not yet exist.
func putNewKV(kv v3.KV, key, val string, leaseID v3.LeaseID) (int64, error) {
	cmp := v3.Compare(v3.Version(key), "=", 0)
	req := v3.OpPut(key, val, v3.WithLease(leaseID))
	txnresp, err := kv.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return 0, err
	}
	if !txnresp.Succeeded {
		return 0, ErrKeyExists
	}
	return txnresp.Header.Revision, nil
}
