-- +goose Up
-- +goose StatementBegin

-- users
CREATE TABLE users (
  id bigint PRIMARY KEY,
  username varchar(200),
  first_name varchar(200),
  last_name varchar(200),
  created_at timestamp NOT NULL
);

-- chats
CREATE TABLE chats (
  id bigint PRIMARY KEY,
  command varchar(100) NOT NULL,
  step smallint NOT NULL,
  data jsonb,
  created_at timestamp NOT NULL,
  updated_at timestamp
);

-- products
CREATE TABLE products (
  id serial PRIMARY KEY,
  name varchar(100) NOT NULL,
  label varchar(100) NOT NULL,
  category varchar(100) NOT NULL,
  uri varchar(2048) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp
);

CREATE UNIQUE INDEX idx_products_name ON products(name);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_label_trgm ON products USING gin (label gin_trgm_ops);
CREATE INDEX idx_products_category ON products(category);

-- product_versions
CREATE TABLE product_versions (
  id serial PRIMARY KEY,
  product_id integer NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,
  release_name varchar(100) NOT NULL,
  release_codename varchar(100),
  release_label varchar(100) NOT NULL,
  release_date timestamp,
  version varchar(20),
  version_release_date timestamp,
  version_release_link varchar(2048),
  created_at timestamp NOT NULL
);

CREATE INDEX idx_product_versions_product_id ON product_versions(product_id);
CREATE UNIQUE INDEX idx_product_versions_version_product_id ON product_versions(version, product_id);
CREATE INDEX idx_product_versions_version_release_date ON product_versions(version_release_date);
CREATE INDEX idx_product_versions_created_at ON product_versions(created_at);

-- watch_lists
CREATE TABLE watch_lists (
  id serial PRIMARY KEY,
  chat_id bigint NOT NULL,
  product_id integer NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,
  created_at timestamp NOT NULL
);

CREATE INDEX idx_watch_lists_chat_id ON watch_lists(chat_id);
CREATE UNIQUE INDEX idx_watch_lists_chat_id_product_id ON watch_lists(chat_id, product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE watch_lists;
DROP TABLE product_versions;
DROP TABLE products;
DROP TABLE chats;
DROP TABLE users;
-- +goose StatementEnd
