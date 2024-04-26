app_name = app
config_file = .app.config.dev.yaml

run.go:
	go run ./cmd/$(app_name) -c $(config_file)

docker_postgres_image = postgres:alpine

docker.run.postgres:
	docker run --rm \
		--name postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		$(args) \
		$(docker_postgres_image)
docker.stop.postgres:
	docker stop postgres

docker_minio_image = quay.io/minio/minio

docker.run.minio:
	docker run --rm \
		--name minio \
		-p 9000:9000 \
		-p 9001:9001 \
		$(args) \
		$(docker_minio_image) \
		server /data --console-address ":9001"
docker.stop.minio:
	docker stop minio
