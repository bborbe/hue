TEAMVAULT ?= ~/.teamvault.json

apply:
	bash -c 'source k8s.env && find . -name "*.yaml" | while read -r file; do cat $${file} | teamvault-config-parser -cache=true -teamvault-config="${TEAMVAULT}" -logtostderr -v=0 | kubectl apply --context=$${CLUSTER_CONTEXT} -f -; done'; \

applytest:
	bash -c 'source k8s.env && find . -name "*.yaml" | while read -r file; do cat $${file} | teamvault-config-parser -cache=true -teamvault-config="${TEAMVAULT}" -logtostderr -v=0; done'; \
