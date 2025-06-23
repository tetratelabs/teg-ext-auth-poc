
## How to run

1. Install TEG: Follow the instructions in the [teg documentation](https://docs.tetrate.io/envoy-gateway/installation/quickstart) to install TEG.

Note: because the POC uses the `Backend` resource to connect to the UDS ext-auth service, you need to enable the `Backend` resource when installing TEG. To do this, run the following command:

```bash
cat <<EOF > teg-values.yaml
gateway-helm:
  config:
    envoyGateway:
      extensionApis:
        enableBackend: true
EOF
```

Then install TEG with the `teg-values.yaml` file:

```bash
helm install teg ${REGISTRY}/teg-envoy-gateway-helm \
 --version ${CHART_VERSION} \
 --values teg-values.yaml \
 -n envoy-gateway-system --create-namespace
```

2. Install the POC: Run the following command to install the POC:

```bash
kubectl apply -f manifests
```

Wait for the pods to be ready:

```bash
kubectl wait --for=condition=Available --timeout=5m -n envoy-gateway-system \
    deployment -l gateway.envoyproxy.io/owning-gateway-name=ext-auth-poc
```

3. Test the POC: Run the following commands to test the POC:

Get the Gateway address:

```bash
export GATEWAY_HOST=$(kubectl get gateway/ext-auth-poc -o jsonpath='{.status.addresses[0].value}')
```

You can also port-forward the Gateway to localhost if load balancer is not available:

```bash
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-name=ext-auth-poc -o jsonpath='{.items[0].metadata.name}')
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8080:80 &
export GATEWAY_HOST=127.0.0.1:8080
```

Curl the Gateway with the authorization header "Bearer token1", which is the bearer token for user1:

```bash
curl -v $GATEWAY_HOST  -H "authorization: Bearer token1"
```

In the response, you should see the request headers added by the datadog tracer.

```bash
... omitted

"X-Datadog-Parent-Id": [
   "7274015792240001721"
  ],
  "X-Datadog-Sampling-Priority": [
   "-1"
  ],
  "X-Datadog-Tags": [
   "_dd.p.tid=6858a92d00000000"
  ],
  "X-Datadog-Trace-Id": [
   "16714982415138379296"
  ],

... omitted

```

In the Envoy logs, you should see the traces successfully submitted to the datadog agent:

```bash
k -n envoy-gateway-system logs deploy/envoy-default-ext-auth-poc-953b9de2 |grep datadog

... omitted
[2025-06-23 01:10:39.472][55][debug][tracing] [source/extensions/tracers/datadog/agent_http_client.cc:108] traces successfully submitted to datadog agent
```
