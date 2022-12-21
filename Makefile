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
	-rm $(OUTPUT_PREFIX)/backend
	-rm $(OUTPUT_PREFIX)/frontend
	-rm $(OUTPUT_PREFIX)/cli-be
	-rm $(OUTPUT_PREFIX)/player
	-rm $(OUTPUT_PREFIX)/static
	-rm $(OUTPUT_PREFIX)/views
	-rm $(PROTO_PREFIX)/backend/*.go $(PROTO_PREFIX)/common/*.go
