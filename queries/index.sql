CREATE TABLE users(
id TEXT PRIMARY KEY UNIQUE NOT NULL,
name TEXT DEFAULT "",
email TEXT DEFAULT "",
password TEXT DEFAULT "", 
is_paid_user boolean not null default false);