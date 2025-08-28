create table if not exists transactions (
    id               bigserial primary key,
    transaction_id   int not null,
    transaction_type text not null,
    created_at       timestamptz not null default now()
);
create unique index unique_event_id on transactions(transaction_id);

create table if not exists balance (
    ingredient_name text not null,
    amount          int not null
);
create unique index unique_ingredient on balance(ingredient_name);

insert into balance(ingredient_name, amount)
    values ('salt', 1000000), ('meal', 1000000), ('water', 1000000);
