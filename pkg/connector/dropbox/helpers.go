package dropbox

import (
	"context"
	"io"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func logBody(ctx context.Context, bodyCloser io.ReadCloser) {
	l := ctxzap.Extract(ctx)
	body := make([]byte, 4096)
	bodyCloser.Read(body)
	l.Info("response body: ", zap.String("body", string(body)))
}
