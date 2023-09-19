docker run \
  --name pgsql-dev \
  –rm \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres