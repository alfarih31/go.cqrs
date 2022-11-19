-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event_store
(
    id         BIGSERIAL primary key,
    event_id uuid not null unique ,
    event_name varchar(255) not null ,
    event_data text not null ,
    metadata text ,
    created_at timestamp without time zone not null default now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_store;
-- +goose StatementEnd
