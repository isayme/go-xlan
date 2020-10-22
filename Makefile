.PHONE: devs, devc
devs:
	CONF_FILE_PATH=./config/dev.yaml go run main.go server

devc:
	CONF_FILE_PATH=./config/dev.yaml go run main.go client
