kind-fromscratch:
	kind create cluster --config k8s/kind/kind-local-infra.yml

kind-ns:
	kubectl apply -f k8s/manifests/namespace.yaml --wait;

kindup: kind-ingress kind-ns
kinddown: kind-minio-rm kind-rm-ingress

kind-ingress:
	kubectl apply -f k8s/kind/ingress-nginx.yaml --wait
kind-rm-ingress:
	kubectl delete -f k8s/kind/ingress-nginx.yaml;

kind-neo4j:
	kubectl apply -f k8s/manifests/neo4j_secrets.yaml
	kubectl apply -f k8s/manifests/neo4j.yaml
kind-neo4j-rm:
	kubectl delete -f k8s/manifests/neo4j.yaml

kind-minio:
	kubectl apply -f k8s/manifests/minio_secrets.yaml --wait; \
	kubectl apply -f k8s/manifests/minio-cm.yaml --wait; \
    kubectl apply -f k8s/manifests/minio.yaml --wait; \
    kubectl apply -f k8s/manifests/minio_init.yaml --wait;

kind-minio-rm:
	kubectl delete -f k8s/manifests/minio_secrets.yaml; \
	kubectl apply -f k8s/manifests/minio-cm.yaml; \
    kubectl delete -f k8s/manifests/minio.yaml; \
    kubectl apply -f k8s/manifests/minio_init.yaml --wait;

kind-drawio:
	kubectl apply -f k8s/manifests/drawio.yaml
kind-drawio-rm:
	kubectl delete -f k8s/manifests/drawio.yaml

kind-rm:
	kind delete cluster