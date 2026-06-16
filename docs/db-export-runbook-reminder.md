# Exporting the DB data (Proxmox → LXC → Docker)

How to pull the `emoji_tracker` Postgres data out as CSVs and get them onto a local machine.

**Topology:** local machine → Proxmox host → LXC container (Docker host) → `emoji-bot-db` container

Replace the placeholders:
- `<ctid>` — LXC container ID (find with `pct list` on Proxmox)
- `<pg_user>` — Postgres user (see step 1)
- `<docker-host>` / `<proxmox-ip>` / `<user>` — your hosts and SSH user

---

## 1. Find the Postgres user

The user comes from `stack.env` (`${POSTGRES_USER}`). Read it from the container:

```bash
docker exec emoji-bot-db env | grep POSTGRES
```

## 2. Export every table to its own CSV (inside the db container)

```bash
docker exec emoji-bot-db sh -c '
  mkdir -p /tmp/csv_export
  for t in $(psql -U <pg_user> -d emoji_tracker -At -c \
    "SELECT tablename FROM pg_tables WHERE schemaname='\''public'\'';"); do
    echo "Exporting $t..."
    psql -U <pg_user> -d emoji_tracker -c \
      "COPY public.\"$t\" TO '\''/tmp/csv_export/$t.csv'\'' WITH CSV HEADER"
  done
'
```

> Note: CSVs are **data only** — no schema/indexes/constraints. For a restorable backup use
> `docker exec -t emoji-bot-db pg_dump -U <pg_user> -d emoji_tracker -F c -f /tmp/emoji_tracker.dump` instead.

## 3. Copy the CSVs out of the container onto the Docker host

```bash
docker cp emoji-bot-db:/tmp/csv_export /root/csv_export
tar czf /root/csv_export.tar.gz -C /root csv_export
```

## 4. Pull the archive from the LXC to the Proxmox host

Run on the **Proxmox** console:

```bash
pct pull <ctid> /root/csv_export.tar.gz /tmp/csv_export.tar.gz
```

## 5. Download from Proxmox to local

Run on your **local machine**:

```bash
scp <user>@<proxmox-ip>:/tmp/csv_export.tar.gz .
tar xzf csv_export.tar.gz
```

## 6. Cleanup (remove the temp files)

```bash
# inside the db container
docker exec emoji-bot-db rm -rf /tmp/csv_export

# on the Docker host (LXC)
rm -f /root/csv_export.tar.gz
rm -rf /root/csv_export

# on the Proxmox host
rm -f /tmp/csv_export.tar.gz
```

---

### Quick reference

| Thing | Value |
|---|---|
| DB name | `emoji_tracker` |
| DB container | `emoji-bot-db` |
| Schema exported | `public` |
| Get into the Docker host from Proxmox | `pct enter <ctid>` |
| List LXCs / VMs | `pct list` / `qm list` |
