DOMAIN=localhost
CERT_PATH=/etc/letsencrypt/live/$(DOMAIN)/fullchain.pem
SCHEME?=http
MINIO_PUBLIC_ENDPOINT?=localhost:9000
MODE?=daemon
SERVICE?=
COMPOSE=docker-compose

ifeq ($(shell [ -f $(CERT_PATH) ] && echo yes),yes)
    SCHEME=https
endif

ifeq (${SCHEME}, https)
    MINIO_PUBLIC_ENDPOINT=${DOMAIN}/minio
endif

# ----------------------------
# Release (deploy) compose
# ----------------------------
DEPLOY_COMPOSE = $(COMPOSE) -f deploy/docker-compose.yml

build:
ifeq ($(strip $(SERVICE)),)
	$(DEPLOY_COMPOSE) build --build-arg SCHEME=$(SCHEME) --build-arg MINIO_PUBLIC_ENDPOINT=${MINIO_PUBLIC_ENDPOINT}
else
	$(DEPLOY_COMPOSE) build $(SERVICE) --build-arg SCHEME=$(SCHEME) --build-arg MINIO_PUBLIC_ENDPOINT=${MINIO_PUBLIC_ENDPOINT}
endif

up: build
ifeq ($(strip $(SERVICE)),)
ifeq ($(MODE),daemon)
	$(DEPLOY_COMPOSE) up -d
else
	$(DEPLOY_COMPOSE) up
endif
else
ifeq ($(MODE),daemon)
	$(DEPLOY_COMPOSE) up -d $(SERVICE)
else
	$(DEPLOY_COMPOSE) up $(SERVICE)
endif
endif

down:
ifeq ($(ERASE),yes)
	$(DEPLOY_COMPOSE) down -v
else
	$(DEPLOY_COMPOSE) down
endif

restart: down up

logs:
	$(DEPLOY_COMPOSE) logs -f

ps:
	$(DEPLOY_COMPOSE) ps

lint:
	cd backend && golangci-lint run ./...

run-local-cicd:
	export DOCKER_HOST=unix:///var/run/docker.sock
	act --container-architecture linux/amd64 \
	  --env DOCKER_HOST=unix:///var/run/docker.sock

# ----------------------------
# Tests compose
# ----------------------------
TEST_COMPOSE = $(COMPOSE) -f tests/docker-compose.yml
WAIT_TIMEOUT = 60

.PHONY: wait integration e2e all unit

wait:
	@end=$$((SECONDS+$(WAIT_TIMEOUT))); \
	while [ $$SECONDS -lt $$end ]; do \
		if curl -s http://localhost:8080/hello; then \
			echo "Приложение готово!"; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "Приложение не успело подняться за $(WAIT_TIMEOUT) секунд" && exit 1

unit:
	$(TEST_COMPOSE) run --rm unit-test

integration:
	$(TEST_COMPOSE) run --rm integration-test

e2e:
	$(TEST_COMPOSE) run --rm e2e-test

all:
	mkdir tests/allure-results || true
	chmod 777 tests/allure-results
	$(MAKE) unit
#	$(TEST_COMPOSE) up -d
#	$(MAKE) wait
	$(MAKE) integration
	$(MAKE) e2e
	$(TEST_COMPOSE) down -v
	$(MAKE) allure-report

allure-report:
	cd tests && \
 		cp -r ./allure-report/history ./allure-results/history 2>/dev/null || : && \
 		allure generate allure-results --clean && allure open allure-report

.PHONY: clean-allure
clean-allure:
	find . -type d -name "allure*" -exec rm -rf {} +\n
