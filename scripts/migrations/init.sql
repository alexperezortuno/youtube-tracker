-- init.sql

CREATE SCHEMA IF NOT EXISTS metrics_db;

\c metrics_db
SET search_path = metrics_db;

CREATE EXTENSION IF NOT EXISTS timescaledb;

load '001_create_tables.sql';
load '002_insert_data.sql';
