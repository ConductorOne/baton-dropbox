package dropbox

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func logBody(ctx context.Context, res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}

	l := ctxzap.Extract(ctx)
	body := make([]byte, 512)
	_, err := res.Body.Read(body)
	if err != nil {
		l.Error("error reading response body", zap.Error(err))
		return
	}
	l.Info("response body: ", zap.String("body", string(body)))
}
