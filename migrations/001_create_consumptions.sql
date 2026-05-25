-- migrations/001_create_consumptions.sql
-- Creates the consumptions table for storing meter energy readings.
-- Compatible with MySQL 5.7+ / MariaDB (XAMPP default).

CREATE DATABASE IF NOT EXISTS bia_energy
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE bia_energy;

CREATE TABLE IF NOT EXISTS consumptions (
    id                  VARCHAR(36)     NOT NULL,
    meter_id            BIGINT          NOT NULL,
    active_energy       DOUBLE          NOT NULL DEFAULT 0,
    reactive_inductive  DOUBLE          NOT NULL DEFAULT 0,
    reactive_capacitive DOUBLE          NOT NULL DEFAULT 0,
    exported            DOUBLE          NOT NULL DEFAULT 0,
    timestamp           DATETIME        NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_meter_id  (meter_id),
    INDEX idx_timestamp (timestamp),
    INDEX idx_meter_ts  (meter_id, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
