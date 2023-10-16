CREATE TABLE IF NOT EXISTS users(
    id TEXT PRIMARY KEY UNIQUE NOT NULL,
    name TEXT DEFAULT "",
    email TEXT DEFAULT "",
    password TEXT DEFAULT "",
    is_paid_user boolean not null default false,
    customer_stripe_id TEXT NOT NULL DEFAULT ""
);

CREATE TABLE IF NOT EXISTS todos(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT "",
    is_complete BOOLEAN DEFAULT FALSE
)