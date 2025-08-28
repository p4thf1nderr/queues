create table if not exists outbox (
    event_id int not null,
    payload BYTEA,
    processed_at timestamptz
);

create table if not exists orders (
    id         bigserial   primary key,
    product_id int not null,
    created_at timestamptz not null default now()
);

