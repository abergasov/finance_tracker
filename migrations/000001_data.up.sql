create table users
(
    u_id             uuid primary key,
    created_at       timestamp,
    updated_at       timestamp,
    email            VARCHAR,
    user_name        VARCHAR,
    default_currency VARCHAR NOT NULL DEFAULT 'USD'
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

create table category_expenses
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID                                    NOT NULL,
    parent_id  BIGINT REFERENCES category_expenses(id) on delete cascade,
    name       TEXT                                    NOT NULL,
    color      TEXT                                    NOT NULL,
    created_at timestamp with time zone default now()  NOT NULL,
    updated_at timestamp with time zone default now()  NOT NULL,
    unique (user_id, parent_id, name)
);

CREATE INDEX category_expenses_user_id_idx ON category_expenses(user_id);
CREATE INDEX category_expenses_parent_id_idx ON category_expenses(parent_id);
CREATE UNIQUE INDEX category_expenses_user_root_uidx ON category_expenses (user_id, name) WHERE (parent_id IS NULL);

CREATE TABLE expense_records
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID        NOT NULL,
    category_id  BIGINT      NOT NULL REFERENCES category_expenses (id) ON DELETE RESTRICT,
    currency     VARCHAR     NOT NULL,
    amount_minor BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX expense_records_user_id_idx ON expense_records (user_id);
CREATE INDEX expense_records_category_id_idx ON expense_records (category_id);
