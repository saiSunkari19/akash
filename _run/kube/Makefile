include ../common.mk
include ../common-commands.mk
include ../common-kind.mk

GATEWAY_HOST     ?= localhost:8080
GATEWAY_ENDPOINT ?= http://$(GATEWAY_HOST)

.PHONY: provider-run
provider-run:
	AKASH_DEPLOYMENT_INGRESS_STATIC_HOSTS="false" \
		$(AKASHCTL) $(KEY_OPTS) provider run \
			--from "$(PROVIDER_KEY_NAME)" \
			--cluster-k8s \
			--gateway-listen-address "$(GATEWAY_HOST)"

.PHONY: provider-status
provider-status:
	$(AKASHCTL) provider status

.PHONY: send-manifest
send-manifest:
	$(AKASHCTL) provider send-manifest "$(SDL_PATH)" \
		--owner     "$(KEY_ADDRESS)" \
		--dseq      "$(DSEQ)"        \
		--gseq      "$(GSEQ)"        \
		--oseq      "$(OSEQ)"        \
		--provider  "$(PROVIDER_ADDRESS)"

.PHONY: provider-lease-status
provider-lease-status:
	$(AKASHCTL) provider status lease-status \
		--owner     "$(KEY_ADDRESS)" \
		--dseq      "$(DSEQ)"        \
		--gseq      "$(GSEQ)"        \
		--oseq      "$(OSEQ)"        \
		--provider  "$(PROVIDER_ADDRESS)"

.PHONY: provider-lease-ping
provider-lease-ping:
	curl -sIH "Host: hello.localhost" localhost:$(KIND_HTTP_PORT)
