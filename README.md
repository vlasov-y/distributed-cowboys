Distributed Cowboys
===

## Getting Started

1. Install [Docker](https://docs.docker.com/engine/install/)
2. Install [Docker Compose](https://docs.docker.com/compose/install/)
3. Launch the database

```shell
docker compose up -d postgres
```

4. Seed the database

```shell
docker compose up seed
```

5. Start cowboys workers

```shell
docker compose up cowboy1 cowboy2 cowboy3
```

6. Enjoy! After it is done, you can restart the stack.

```shell
docker compose up seed
docker compose up cowboy1 cowboy2 cowboy3
```
