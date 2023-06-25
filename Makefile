
.PHONY: subscriber
subscriber:
	go build -o build/subscriber ./cmd/subscriber/subscriber.go

.PHONY: publisher
publisher:
	go build -o build/publisher ./cmd/publisher/publisher.go

.PHONY: subscriberd
subscriberd:
	go build -o build/subscriberd ./cmd/bootstrap-subscriber/bootstrap-subscriber.go

.PHONY: publisherd
publisherd:
	go build -o build/publisherd ./cmd/bootstrap-publisher/bootstrap-publisher.go