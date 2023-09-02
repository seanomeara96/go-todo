CREATE TABLE IF NOT EXISTS users(
id TEXT PRIMARY KEY UNIQUE NOT NULL,
name TEXT DEFAULT "",
email TEXT DEFAULT "",
password TEXT DEFAULT "", 
is_paid_user boolean not null default false);

ALTER TABLE users ADD COLUMN customer_stripe_id TEXT NOT NULL DEFAULT "";