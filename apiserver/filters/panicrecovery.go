package filters

import (
	"net/http"

	"github.com/golang/glog"

	utilruntime "denggotech.cn/heque/heque/util/runtime"
)

// WithPanicRecovery wraps an http Handler to recover and log panics.
func WithPanicRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer utilruntime.HandleCrash(func(err interface{}) {
			glog.Errorf("servlet panic'd on %v %v", r.Method, r.RequestURI)
			http.Error(w, "This request caused servlet to panic. Look in the logs for details.", http.StatusInternalServerError)
		})
		// Dispatch to the internal handler
		handler.ServeHTTP(w, r)
	})
}
