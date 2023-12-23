build:
	docker build -t brent .

run: build
	docker run $(DOCKER_ARGS) --rm -p 8989:9080 -it -v ${HOME}/.kube:/root/.kube brent --https-listen-port 0

run-host: build
	docker run $(DOCKER_ARGS) --net=host --uts=host --rm -it -v ${HOME}/.kube:/root/.kube brent --kubeconfig /root/.kube/config --http-listen-port 8989 --https-listen-port 0
