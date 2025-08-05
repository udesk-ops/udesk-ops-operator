/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

var _ = Describe("APIServer", func() {
	var (
		server     *APIServer
		fakeClient client.Client
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		err := opsv1beta1.AddToScheme(testScheme)
		Expect(err).NotTo(HaveOccurred())

		fakeClient = fake.NewClientBuilder().
			WithScheme(testScheme).
			Build()

		server = NewAPIServer(fakeClient, ":8080")
	})

	Describe("NewAPIServer", func() {
		Context("when creating a new server", func() {
			It("should create server with correct address", func() {
				Expect(server).NotTo(BeNil())
				Expect(server.addr).To(Equal(":8080"))
			})

			It("should have a valid router", func() {
				Expect(server.router).NotTo(BeNil())
			})

			It("should have a valid client", func() {
				Expect(server.client).NotTo(BeNil())
			})
		})
	})

	Describe("Router Setup", func() {
		Context("when setting up routes", func() {
			It("should setup routes without error", func() {
				// Just test that setupRoutes doesn't panic
				Expect(func() {
					server.setupRoutes()
				}).NotTo(Panic())
			})
		})
	})

	Describe("Server Lifecycle", func() {
		Context("when starting server", func() {
			It("should start server successfully", func(ctx SpecContext) {
				serverCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
				defer cancel()

				// Start server (will auto-stop when context is cancelled)
				err := server.Start(serverCtx)
				// Context cancellation is expected, not an error
				if err != nil && err != context.DeadlineExceeded {
					Fail(fmt.Sprintf("Unexpected server error: %v", err))
				}
			}, SpecTimeout(time.Second*2))
		})
	})

	Describe("Router Configuration", func() {
		Context("when router is created", func() {
			It("should have a router instance", func() {
				Expect(server.router).NotTo(BeNil())
			})

			It("should be able to handle unknown routes", func() {
				req := httptest.NewRequest("GET", "/unknown", nil)
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)

				// Should return 404 for unknown routes
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
