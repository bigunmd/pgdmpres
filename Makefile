app_name = app
config_file = .app.config.dev.yaml

run.go:
	go run ./cmd/$(app_name) -c $(config_file)

docker_postgres_image = postgres:alpine

docker.run.postgres:
	docker run --rm -d \
		--name postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=postgres \
		-p 5432:5432 \
		-v ~/postgres/data:/var/lib/postgresql/data \
		$(args) \
		$(docker_postgres_image)
docker.stop.postgres:
	docker stop postgres

docker_minio_image = quay.io/minio/minio

docker.run.minio:
	docker run --rm -d \
		--name minio \
		-p 9000:9000 \
		-p 9001:9001 \
		-v ~/minio/data:/data \
		$(args) \
		$(docker_minio_image) \
		server /data --console-address ":9001"
docker.stop.minio:
	docker stop minio

image_tag = test

docker.build.app:
	docker buildx build . -f docker/Dockerfile -t $(app_name):$(image_tag) --load

docker.compose.build:
	docker compose -f docker/docker-compose.yaml build
docker.compose.up:
	docker compose -f docker/docker-compose.yaml up
docker.compose: docker.compose.build docker.compose.up