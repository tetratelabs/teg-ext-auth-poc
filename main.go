// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var udsAddr = "/var/run/envoy-uds/ext-auth.sock"

func main() {
	if _, err := os.Stat(udsAddr); err == nil {
		if err := os.RemoveAll(udsAddr); err != nil {
			log.Fatalf("failed to remove: %v", err)
		}
	}

	listener, err := net.Listen("unix", udsAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	users := TestUsers()

	// Define an HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Check(w, r, users)
	})

	// Start the ext auth HTTP server
	log.Printf("Ext auth server is listening on %s\n", udsAddr)
	go func() {
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Ext auth server failed: %v\n", err)
		}
	}()

	// Start the health check HTTP server
	http.HandleFunc("/healthz", healthCheckHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func Check(w http.ResponseWriter, r *http.Request, users Users) {
	authorization := r.Header["Authorization"]
	log.Println(r.Header)
	log.Println(authorization)

	if len(authorization) > 0 {
		extracted := strings.Fields(authorization[0])
		if len(extracted) == 2 && extracted[0] == "Bearer" {
			valid, user := users.Check(extracted[1])
			if valid {
				w.Header().Add("x-current-user", user)
				w.WriteHeader(http.StatusOK)
			}
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
}

// Users holds a list of users.
type Users map[string]string

// Check checks if a key could retrieve a user from a list of users.
func (u Users) Check(key string) (bool, string) {
	value, ok := u[key]
	if !ok {
		return false, ""
	}
	return ok, value
}

func TestUsers() Users {
	return map[string]string{
		"token1": "user1",
		"token2": "user2",
		"token3": "user3",
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", udsAddr)
				},
		},
	}

	req, err := http.NewRequest("GET", "http://unix/", nil)
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("authorization", "Bearer token1")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not check: %v", err)
	}
	if resp != nil && resp.StatusCode == http.StatusOK {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
