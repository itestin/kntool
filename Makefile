.PHONY: image
image:
	docker build -f webhook.Dockerfile -t zhaihuailou/kntool:latest .
	docker build -f sidecar.Dockerfile -t zhaihuailou/kntool-sidecar:latest .
.PHONY: push
push:
	docker push zhaihuailou/kntool:latest
	docker push zhaihuailou/kntool-sidecar:latest
.PHONY: tar
tar:
	docker save push zhaihuailou/kntool:latest -o kntool.tar
	docker save push zhaihuailou/kntool-sidecar:latest -o kntool-sidecar.tar
.PHONY: load
load:
	docker save push zhaihuailou/kntool:latest -o kntool.tar
	docker save push zhaihuailou/kntool-sidecar:latest -o kntool-sidecar.tar
.PHONY: deploy
deploy:
	kubectl apply -f deploy/deployment.yml

