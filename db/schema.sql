-- noinspection SqlNoDataSourceInspectionForFile

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS event_streams (
    sequence SERIAL PRIMARY KEY,
    aggregate VARCHAR(50)[] NOT NULL,
    type VARCHAR(50) NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    revision SMALLINT NOT NULL DEFAULT 1,
    payload JSONB
);

CREATE INDEX IF NOT EXISTS event_streams_aggregate_idx ON event_streams USING GIN(aggregate);

CREATE TABLE IF NOT EXISTS subscribers (
    id VARCHAR(100) PRIMARY KEY,
    last_acknowledged_sequence INTEGER DEFAULT 0,
    last_seen_at TIMESTAMP
);
