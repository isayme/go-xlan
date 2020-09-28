.PHONE: devs, devc
devs:
	CONF_FILE_PATH=./config/dev.yaml go run cmd/server/main.go

devc:
	CONF_FILE_PATH=./config/dev.yaml go run cmd/client/main.go