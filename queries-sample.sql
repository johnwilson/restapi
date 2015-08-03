-- name: drop-database
DROP DATABASE `mydb`;
-- name: create-db
CREATE DATABASE mydb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- name: create-user-table
CREATE TABLE user (
	id INT AUTO_INCREMENT PRIMARY KEY,
	username VARCHAR(100) UNIQUE,
	firstname VARCHAR(100),
	lastname VARCHAR(100),
	email VARCHAR(150),
	active TINYINT(1),
	password VARBINARY(255),
	security_role_id INT,
	created INT,
	updated INT
);