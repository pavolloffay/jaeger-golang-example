FROM scratch

EXPOSE 8080/tcp

ENV JAEGER_SERVICE_NAME=golang-app-from-process \
  JAEGER_SAMPLER_TYPE=const \
  JAEGER_ENDPOINT="http://jaeger-collector.istio-system.svc:14268/api/traces?format=jaeger.thrift" \
  JAEGER_SAMPLER_PARAM=1

COPY app /go/bin/

ENTRYPOINT ["/go/bin/app"]
