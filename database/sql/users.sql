CREATE table users (
  id SERIAL primary key,
  username text unique not null,
  password text default null,
  ad boolean not null,
  namespaces text[],
  admin boolean default false,
  created_at date default current_date,
  deleted_at date default null
);
CREATE EXTENSION pgcrypto;
