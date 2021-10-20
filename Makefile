.PHONY: compile
compile-protoc:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        pkg/protobuf/v1/task.proto

compile-thrift:
	 thrift -r -out pkg/thrift --gen go pkg/thrift/tasks.thrift

generate-sql:
	sqlc generate

dev-sql:
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec db psql -U postgres

generate-mock:
	mockery --name taskHandler --keeptree --exported --recursive

dev-up:
	@docker-compose -f deployments/ci/docker-compose.yaml up --build