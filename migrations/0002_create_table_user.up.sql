create table if not exists "users" (
    id UUID primary key default gen_random_uuid(),
    name varchar(50) not null,
    email varchar(100) not null unique,
    password text not null,
    created_at timestamptz not null default now()
);
