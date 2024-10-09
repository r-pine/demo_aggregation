# Rpine Demo aggregator

## Run:


### Start nginx proxy:

```bash
docker compose -f docker-compose.ssl.yml up --build -d
```


### Start redis:

```bash
docker compose -f docker-compose.db.yml up --build -d
```


### Start aggregator:

```bash
docker compose -f docker-compose.yml up --build -d
```