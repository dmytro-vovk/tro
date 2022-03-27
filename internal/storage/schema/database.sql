CREATE DATABASE IF NOT EXISTS tro;

USE tro;

CREATE TABLE IF NOT EXISTS operators (
    id    INT AUTO_INCREMENT PRIMARY KEY,
    login VARCHAR(20) NOT NULL,
    CONSTRAINT operators_login_uindex UNIQUE (login)
);

