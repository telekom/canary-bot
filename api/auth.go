package api

import (
	"context"
	"errors"

	connect "github.com/bufbuild/connect-go"
)

func (a *Api) NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(
			func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				authToken := req.Header().Get("Authorization")
				// check if token is set
				if authToken == "" {
					a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed")
					return nil, connect.NewError(
						connect.CodeUnauthenticated,
						errors.New("no token provided"),
					)
				}

				// check if token is correct
				for _, t := range a.config.Tokens {
					if authToken[7:] == t {
						a.log.Infow("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "succeded")
						return next(ctx, req)
					}
				}
				a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed")
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("auth failed"),
				)
			})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
