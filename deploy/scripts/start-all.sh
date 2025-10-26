#!/bin/bash

docker-compose -f docker-compose.infra.yaml up -d
docker-compose -f docker-compose.monitoring.yaml up -d

docker-compose up -d gateway
ENV_FILE=.env.ro1 docker-compose --env-file .env.ro1 -p quickflow_ro1 up -d gateway
ENV_FILE=.env.ro2 docker-compose --env-file .env.ro2 -p quickflow_ro2 up -d gateway

ENV_FILE=env-mirror/.env.mirror docker-compose --env-file env-mirror/.env.mirror -p quickflow_mirror up -d gateway
ENV_FILE=env-mirror/.env.mirror.ro1 docker-compose --env-file env-mirror/.env.mirror.ro1 -p quickflow_mirror-ro1 up -d gateway
ENV_FILE=env-mirror/.env.mirror.ro2 docker-compose --env-file env-mirror/.env.mirror.ro2 -p quickflow_mirror-ro2 up -d gateway
