generate-api:
	goctl api go -api services/gateway/api/contract/watchProgress.api -dir services/gateway/api/

docs-generate:
	rm -rf services/gateway/api/docs/watchProgress.json
	goctl api swagger --api services/gateway/api/contract/watchProgress.api --dir services/gateway/api/docs/

test:
	docker compose -f docker-compose.test.yaml up -d
	go test -v -tags integration $$(go list ./... | grep -Ev '/constants|/third_party')
	docker compose -f docker-compose.test.yaml down

start-docker:
	docker compose -f docker-compose.yaml up -d

stop-docker:
	docker compose -f docker-compose.yaml down

run:
	go run services/gateway/api/watchprogress.go