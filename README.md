# Emoji Counter Bot for Discord

Go bot that counts emoji usage in a Discord server. Runs as three containers: the bot, Postgres, and Redis.

## Local development

1. Copy `.env.example` to `.env` and fill in the values (Discord token, passwords, etc).
2. Bring it up:
   ```shell
   docker compose up -d --build
   ```
   `compose.override.yaml` is picked up automatically and republishes Postgres on `localhost:5433` and Redis on `localhost:6379` so you can connect with a GUI.
3. Tear down:
   ```shell
   docker compose down
   ```

## Deploying via Portainer

The same `compose.yaml` works as a Portainer **stack**. The override file is ignored in production, so DB/Redis stay on the internal network only.

1. In Portainer: **Stacks → Add stack → Repository**, point it at this repo.
2. Set the compose path to `compose.yaml` (do **not** include `compose.override.yaml`).
3. Add the environment variables from `.env.example` in the stack's *Environment variables* section.
4. Deploy.

To redeploy after a code change, hit **Pull and redeploy** in the Portainer stack view.

## Containers

| Service | Image                | Notes                                          |
|---------|----------------------|------------------------------------------------|
| backend | built from `backend/`| The bot. Memory-capped at 256M.                |
| db      | postgres:16-alpine   | Initial schema seeded from `migration.sql`.    |
| redis   | redis:7.4.2-alpine   | 100M LRU cache, password-protected.            |

The backend Dockerfile is a multi-stage build that produces a minimal Alpine-based runtime image.
