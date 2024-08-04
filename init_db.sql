CREATE TABLE IF NOT EXISTS Users (
    id        INTEGER PRIMARY KEY,
    name      TEXT UNIQUE NOT NULL,
    auth_hash TEXT NOT NULL,
    auth_salt TEXT NOT NULL,
    admin     INTEGER DEFAULT 0 NOT NULL
);

