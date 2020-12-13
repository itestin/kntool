image:
	docker build -f webhook.Dockerfile -t zhaihuailou/kntool:latest .
	docker build -f sidecar.Dockerfile -t zhaihuailou/kntool-sidecar:latest .

push:
	docker push zhaihuailou/kntool:latest
	docker push zhaihuailou/kntool-sidecar:latest

tar:
	docker save push zhaihuailou/kntool:latest -o kntool.tar
	docker save push zhaihuailou/kntool-sidecar:latest -o kntool-sidecar.tar

load:
	docker save push zhaihuailou/kntool:latest -o kntool.tar
	docker save push zhaihuailou/kntool-sidecar:latest -o kntool-sidecar.tar

deploy:
	kubectl apply -f deploy/deployment.yml

.PHONY: image, push
