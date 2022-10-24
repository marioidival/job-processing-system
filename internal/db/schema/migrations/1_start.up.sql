begin;

create extension if not exists "uuid-ossp";

create type JOB_STATUS as enum ('PENDING', 'PROCESSED', 'ERROR');

create table if not exists jobs(
    id serial primary key,
    job_uuid uuid default gen_random_uuid() not null,
    status JOB_STATUS default 'PENDING' not null,
    data integer [],
    result integer,
    action varchar not null,
    created_at timestamp without time zone default timezone('utc'::text, now()) not null,
    updated_at timestamp without time zone default timezone('utc'::text, now()) not null
);

create table if not exists config(
    time integer not null
);

commit;