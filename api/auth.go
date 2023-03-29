/*
 * canary-bot
 *
 * (C) 2022, Maximilian Schubert, Deutsche Telekom IT GmbH
 *
 * Deutsche Telekom IT GmbH and all other contributors /
 * copyright owners license this file to you under the Apache
 * License, Version 2.0 (the "License"); you may not use this
 * file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package api

import (
	"context"
	"errors"
	"net/http"

	connect "github.com/bufbuild/connect-go"
)

// http auth handler
func (a *Api) NewAuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			a.log.Warnw("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// check if token is correct
		for _, t := range a.config.Tokens {
			if authToken[7:] == t {
				a.log.Infow("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "succeded")
				h.ServeHTTP(w, r)
				return
			}
		}
		a.log.Warnw("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}

// grpc auth interceptor
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
