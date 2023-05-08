/* allows the use of the uuid type */
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

/* helps with creating partial indexes for LIKE queries */
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT uuid_generate_v4 (),
    first_name VARCHAR (255) NOT NULL,
    last_name VARCHAR (255) NOT NULL,
    nickname VARCHAR (255) NOT NULL CHECK (nickname <> ''),
    password VARCHAR (255) NOT NULL CHECK (password <> ''),
    email VARCHAR (255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT id_unique UNIQUE (id),
    CONSTRAINT email_unique UNIQUE (email)
);

/* first name index */
CREATE INDEX idx_users_first_name ON users (first_name);

/* last name index */
CREATE INDEX idx_users_last_name ON users (last_name);

/* email index */
CREATE UNIQUE INDEX email_idx ON users (email);