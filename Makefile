runluabot: $(shell find . -name "*.go")
	go build -ldflags="-s -w" -o ./runluabot

deploy: runluabot
	ssh root@nusakan-58 'systemctl stop runluabot'
	scp runluabot nusakan-58:runluabot/runluabot
	ssh root@nusakan-58 'systemctl start runluabot'
