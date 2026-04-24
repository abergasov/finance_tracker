create table users
(
    u_id         uuid primary key,
    created_at   timestamp,
    updated_at   timestamp,
    email        varchar,
    user_name    varchar
);

create unique index if not exists users_email_uidx on users (lower(email));
