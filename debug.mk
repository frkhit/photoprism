.PHONY: docker-start
docker-start:
	cd docker/examples/sqlite/ && docker-compose up -d

.PHONY: rebuild-docker
rebuild-docker:
	docker build -t  photoprism_release \
	-f Dockerfile.photoprism \
	.
