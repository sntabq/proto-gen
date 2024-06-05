CREATE TABLE IF NOT EXISTS auth.tokens
(
    hash    bytea PRIMARY KEY,
    user_id BIGINT                      NOT NULL REFERENCES auth.users ON DELETE CASCADE ,
    expiry  TIMESTAMP(0) WITH TIME ZONE NOT NULL
);