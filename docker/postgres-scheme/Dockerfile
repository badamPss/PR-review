FROM postgres:15-alpine

RUN apk update && apk add --no-cache bash postgresql-client

ENV PGHOST=postgres \
    PGPORT=5432 \
    PGUSER=postgres \
    PGPASSWORD=password \
    PGDATABASE=pr_review \
    PGCONNECT_TIMEOUT=5 \
    WAIT_TIMEOUT_SEC=60

WORKDIR /app

COPY scripts/migrate.sh /app/migrate.sh
COPY db/migration /app/migration

RUN chmod a+x /app/migrate.sh

CMD ["sh", "-c", "MIGRATIONS_DIR=/app/migration /app/migrate.sh"]


