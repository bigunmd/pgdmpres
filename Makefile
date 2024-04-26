app_name = app
config_file = .app.config.dev.yaml

run.go:
	go run ./cmd/$(app_name) -c $(config_file)