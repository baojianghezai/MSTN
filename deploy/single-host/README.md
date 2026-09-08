# Single-host deployment

This deployment runs the public recruitment frontend, API, MySQL, and Redis on one host. Only Nginx port `80` is published. MySQL, Redis, and the Go API stay on the internal Docker network.

## Prepare the host

- Docker Engine 25+ with Docker Compose v2
- At least 2 vCPU, 4 GB RAM, and persistent disk storage
- Open only TCP `80` in the cloud security group and host firewall

Copy this directory to the server, then create the two local secret files:

```bash
cp .env.example .env
cp config.production.example.yaml config.production.yaml
```

Set unique random values in `.env` and `config.production.yaml`. The following values must agree between the files:

| `.env` | `config.production.yaml` |
| --- | --- |
| `MYSQL_DATABASE` | `mysql.db-name` |
| `MYSQL_USER` | `mysql.username` |
| `MYSQL_PASSWORD` | `mysql.password` |
| `REDIS_PASSWORD` | `redis.password` |

Generate secrets, for example:

```bash
openssl rand -base64 36
```

Start the services from this directory:

```bash
docker compose up -d --build
docker compose ps
curl http://127.0.0.1/health
```

The first startup creates tables and seeds initial data. After confirming the migration and taking a database backup, change `system.disable-auto-migrate` in `config.production.yaml` to `true`, then restart the API:

```bash
docker compose up -d server
```

## Operations

View logs with `docker compose logs -f server` and back up the database before every application upgrade. Do not run `docker compose down -v` on a production host: it deletes the database, Redis, upload, and log volumes.

This deployment is HTTP-only while there is no domain. When a domain is available, put a TLS reverse proxy in front of `web` or switch the public entry to an HTTPS-capable proxy. Keep `/api/v1/` unchanged when proxying to the API.

Payment remains disabled by design. Do not enable it until the business supplies a public HTTPS domain, WeChat/Alipay merchant credentials, certificates, and callback verification has been completed.
