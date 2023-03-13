package apiserver

import (
	"denggotech.cn/heque/heque/apiserver/handlers"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang/glog"
)

// APIServer is the main servlet implementation.
type APIServer struct {
	cfg *Config
}

// NewAPIServer creates and initializes a new APIServer object.
func NewAPIServer(cfg *Config) *APIServer {
	s := &APIServer{
		cfg: cfg,
	}

	ws := new(restful.WebService)
	ws.Route(ws.POST("/registry/jobs/{queue}").To(func(req *restful.Request, res *restful.Response) {
		handlers.EnqueueJobHandler(cfg.Storage, cfg.Prefix, req, res)
	}).
		Doc("enqueue a job into specified queue"))
	ws.Route(ws.GET("/registry/jobs/{queue}").To(func(req *restful.Request, res *restful.Response) {
		handlers.DequeueJobHandler(cfg.Storage, cfg.Prefix, req, res)
	}).
		Doc("dequeue a job into specified queue"))

	restful.DefaultContainer.Add(ws)

	return s
}

// ListenAndServe runs the servlet HTTP server.
func (s *APIServer) ListenAndServe(addr string) error {
	glog.Infof("Serving HTTP at http://%s", addr)
	return http.ListenAndServe(addr, nil)
}
