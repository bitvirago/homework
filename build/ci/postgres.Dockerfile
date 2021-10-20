FROM postgres:11.11-alpine
COPY pkg/sql/schema/schema.sql /docker-entrypoint-initdb.d/
