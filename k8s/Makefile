# MIGRATED to nuke-prod 2026-08-21.
# Also replaced `teamvault-config-parser` (no such binary any more) with the
# current `teamvault-cli config parse`, matching every quant/trading Makefile.
# NOTE: the previous recipe piped into `kubectlquant` from inside `bash -c`.
# kubectlquant/kubectlnuke* are ZSH FUNCTIONS with no executable anywhere on PATH,
# so they are undefined in bash and that pipe could not have worked. Use an
# explicit KUBECONFIG with the real kubectl binary instead.
TEAMVAULT ?= ~/.teamvault.json

apply:
	bash -c 'source k8s.env && find . -name "*.yaml" | while read -r file; do cat $${file} | teamvault-cli config parse | KUBECONFIG=$$HOME/.kube/nuke-prod kubectl apply -f -; done'; \

applytest:
	bash -c 'source k8s.env && find . -name "*.yaml" | while read -r file; do cat $${file} | teamvault-cli config parse; done'; \
