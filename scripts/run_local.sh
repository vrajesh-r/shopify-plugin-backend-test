docker pull postgres
docker run --rm --name pg-docker -e POSTGRES_PASSWORD=mooncakes -d -p 5432:5432 -v $HOME/docker/volumes/postgres:/var/lib/postgresql/data postgres
psql -h 127.0.0.1 -p 5432 -U postgres