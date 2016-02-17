SHELL := /bin/bash

check-docker:
	@if [ -z $$(which docker) ]; then \
		echo "Missing \`docker\` client which is required for development"; \
		exit 2; \
	fi

check-kubectl:
	@if [ -z $$(which kubectl) ]; then \
		echo "Missing \`kubectl\` client which is required for development"; \
		exit 2; \
	fi
