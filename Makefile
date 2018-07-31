
.PHONY: dep
dep:
	dep ensure

.PHONY: fmt
fmt:
	go fmt ./cmd

.PHONY: build
build:
	CGO_ENABLED=0 installsuffix=cgo go build -o ./cmd/app ./cmd/main.go

.PHONY: build-docker
build-docker: build
	docker build -t golang-app cmd

.PHONY: oc-deploy
oc-deploy:
	oc adm policy add-scc-to-user privileged -z default -n myproject
	istioctl kube-inject -f ./kubernetes/deployment.yml | oc apply -f -
	oc create -f ./kubernetes/service.yml
	oc expose service golang-app

.PHONY: oc-delete
oc-delete:
	oc delete all,service -l app=golang-app
	oc delete route/golang-app

.PHONY: oc-logs
oc-logs:
	oc logs deploy/golang-app -c golang-app

.PHONY: curl
curl:
	curl golang-app-myproject.`minishift ip`.nip.io/chaining
