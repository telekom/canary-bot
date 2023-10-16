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
	"strings"

	connect "github.com/bufbuild/connect-go"
)

// NewAuthHandler returns a handler for HTTP authorization
func (a *Api) NewAuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		splitToken := strings.Split(r.Header.Get("Authorization"), "Bearer")
		// check if token is set
		if len(splitToken) != 2 {
			a.log.Warnw("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "failed", "reason", "no bearer token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// get token
		authToken := strings.TrimSpace(splitToken[1])

		// check if token is correct
		for _, t := range a.config.Tokens {
			if authToken == t {
				a.log.Infow("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "succeeded")
				h.ServeHTTP(w, r)
				return
			}
		}
		a.log.Warnw("Request", "host", r.Header.Get("X-Forwarded-Host"), "auth", "failed", "reason", "invalid token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}

// NewAuthInterceptor returns grpc auth interceptor to handle authorization
func (a *Api) NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			splitToken := strings.Split(req.Header().Get("Authorization"), "Bearer")
			// check if token is set
			if len(splitToken) != 2 {
				a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed", "reason", "no bearer token")
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}

			// get token
			authToken := strings.TrimSpace(splitToken[1])

			// check if token is correct
			for _, t := range a.config.Tokens {
				if authToken == t {
					a.log.Infow("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "succeeded")
					return next(ctx, req)
				}
			}
			a.log.Warnw("Request", "host", req.Header().Get("X-Forwarded-Host"), "auth", "failed", "reason", "invalid token")
			return nil, connect.NewError(
				connect.CodeUnauthenticated,
				errors.New("auth failed"),
			)
		}
	}
	return interceptor
}
