create table users
(
    u_id         uuid primary key,
    created_at   timestamp,
    updated_at   timestamp,
    email        varchar,
    user_name    varchar
);

create unique index if not exists users_email_uidx on users (lower(email));

CREATE TABLE currency_rates (
    id            BIGSERIAL PRIMARY KEY,
    timestamp     TIMESTAMPTZ NOT NULL,
    timestamp_num INTEGER NOT NULL,
    rates         JSONB NOT NULL DEFAULT '{}'::jsonb
);


CREATE UNIQUE INDEX currency_rates_timestamp_num_uq ON currency_rates (timestamp_num);
CREATE INDEX currency_rates_timestamp_idx ON currency_rates (timestamp DESC);
