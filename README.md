# Rpine Demo aggregator

Presentation https://docs.google.com/presentation/d/1-i9OxO3UkjYlabVK7K7cBuhugQtjUhOba957LamjiW0/edit#slide=id.p

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