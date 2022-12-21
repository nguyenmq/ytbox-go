OUTPUT_PREFIX = bin
PROTO_PREFIX = internal/proto

proto: $(PROTO_PREFIX)/backend/*.proto $(PROTO_PREFIX)/common/*.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTO_PREFIX)/common/*.proto $(PROTO_PREFIX)/backend/*.proto

backend:
	go build -o $(OUTPUT_PREFIX)/$@ ./cmd/ytb-be

frontend:
	go build -o $(OUTPUT_PREFIX)/$@ ./cmd/ytb-fe

frontend-assets: frontend
	-ln -s ../internal/frontend/static $(OUTPUT_PREFIX)/static
	-ln -s ../internal/frontend/views $(OUTPUT_PREFIX)/views

cli-be:
	go build -o $(OUTPUT_PREFIX)/$@ ./cmd/ytb-be-cli

player:
	go build -o $(OUTPUT_PREFIX)/$@ ./cmd/ytb-player

bins: backend frontend cli-be player

all: proto backend frontend cli-be player frontend-assets

clean:
	-rm $(OUTPUT_PREFIX)/backend $(OUTPUT_PREFIX)/frontend $(OUTPUT_PREFIX)/cli-be $(OUTPUT_PREFIX)/player $(OUTPUT_PREFIX)/static $(OUTPUT_PREFIX)/views
	-rm $(PROTO_PREFIX)/backend/*.go $(PROTO_PREFIX)/common/*.go
