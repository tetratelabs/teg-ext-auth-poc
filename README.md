# TEG Ext Auth POC

This POC is a simple example of how to use the Ext Auth to authenticate users and reroute them to backend services based
on the appended headers in the authentication response.

## Architecture

The architecture of this POC is as follows:
* TEG: TEG configures the Envoy proxy to use the Ext Auth filter to authenticate users. TEG also configures the Envoy proxy to route requests with the "X-Current-User" header equal to "user1" to the backend service v1 and requests with the "X-Current-User" header equal to "user2" to the backend service v2.
* Envoy: The Envoy proxy that routes requests to the backend services. The Envoy proxy communicates with the Ext Auth service using a Unix Domain Socket (UDS) to authenticate users.
* Ext Auth: A simple service deployed as a sidecar to the Envoy proxy that authenticates users and appends a "X-Current-User" header to the request based on the user. The allowed users are "token1
:user1","token2:user2", "token3:user3".
* Backend: Two versions of a simple backend service that returns a simple response with the request headers and the service pod name.

```
                                ┌──────────────────────────────┐
                                │                              │
                                │                              │
                                │             TEG              │
                                │                              │
                                │                              │
                                └──────────────┬───────────────┘
                                               │                                 ┌────────────────┐
                                               │                                 │                │
                                               │  Configuration                  │                │
                                               │                                 │                │
                                               ▼                                 │                │
                                 ┌─────────────────────────────┐                 │    Backend     │
                                 │                             │         ┌───────►       V1       │
                                 │             Pod             │         │       │                │
                                 │   ┌──────────────────────┐  │         │       │                │
                                 │   │                      │  │         │       │                │
                                 │   │                      │  │         │       └────────────────┘
curl $GATEWAY_HOST \             │   │                      │  │         │
-H "authorization: Bearer token1"│   │        Envoy         │  │  user1  │
                                 │   │                      │  ┼─────────┘
  ──────────────────────────────►│   │                      │  │
                                 │   │                      │  │  user2
                                 │   │                      │  ┼──────────┐
                                 │   └──────────┬───────────┘  │          │
                                 │              │              │          │
                                 │              │ UDS          │          │      ┌────────────────┐
                                 │   ┌──────────▼───────────┐  │          │      │                │
                                 │   │                      │  │          │      │                │
                                 │   │       Ext Auth       │  │          │      │                │
                                 │   │                      │  │          └─────►│                │
                                 │   └──────────────────────┘  │                 │    Backend     │
                                 └─────────────────────────────┘                 │       V2       │
                                                                                 │                │
                                                                                 │                │
                                                                                 │                │
                                                                                 └────────────────┘
```

## How to run

1. Install TEG: Follow the instructions in the [teg documentation](https://docs.tetrate.io/envoy-gateway/installation/quickstart) to install TEG.

2. Install the POC: Run the following command to install the POC:

```bash
kubectl apply -f manifests
```

3. Test the POC: Run the following commands to test the POC:

Get the Gateway address:

```bash
export GATEWAY_HOST=$(kubectl get gateway/ext-auth-poc -o jsonpath='{.status.addresses[0].value}')
```

Curl the Gateway with the authorization header "Bearer token1", which is the bearer token for user1:

```bash
curl -v $GATEWAY_HOST  -H "authorization: Bearer token1"
```

In the response, you should see:

* A "X-Current-User" header with the value "user1" was appended to the request by the Ext Auth service.
* The request was routed to the backend service v1.

```bash
... omitted
"X-Current-User": [
   "user1"
  ],

... omitted

"pod": "backend-app-v1-889987d96-hmdvs"
```

Curl the Gateway with with the authorization header "Bearer token2", which is the bearer token for user2:

```bash
curl -v $GATEWAY_HOST  -H "authorization: Bearer token2"
```

In the response, you should see:

* A "X-Current-User" header with the value "user2" was appended to the request by the Ext Auth service.
* The request was routed to the backend service v2.

```bash
... omitted
"X-Current-User": [
   "user2"
  ],

... omitted

"pod": "backend-app-v2-7d6658bc9-47ng5"
```

If you curl the Gateway with the authorization header "Bearer token3":

```bash
curl -v $GATEWAY_HOST  -H "authorization: Bearer token3"
```

In the response, you should see the request was randomly routed to either the backend service v1 or v2, as the HTTPRoute
does not have a defined route for user3, and the request is routed to the default backend service.

If you curl the Gateway with the authorization header "Bearer token4":

```bash
curl -v $GATEWAY_HOST  -H "authorization : Bearer token4"
```

In the response, you should see a 403 Forbidden response, as the Ext Auth service does not recognize the token.
