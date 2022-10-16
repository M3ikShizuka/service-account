ms-cluster-name := cluster-service-account
kind-config-loadbalancer-path := ../deployments/kubernetes/kind/kind-cluster-loadbalancer-ingres-test.yaml
###########################################
# Test
###########################################
run:
	# loadbalancer
	kind create cluster --name ${ms-cluster-name} --config ${kind-config-loadbalancer-path}

	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	kubectl wait --namespace ingress-nginx \
	--for=condition=ready pod \
	--selector=app.kubernetes.io/component=controller \
	--timeout=90s
	kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/usage.yaml
	
	# should output "foo"
	curl localhost/foo
	# should output "bar"
	curl localhost/bar

stop:
	kind delete cluster --name ${ms-cluster-name}