-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event_stream
(
    id         BIGSERIAL primary key,
    stream_id varchar(255) not null ,
    stream_version INTEGER not null DEFAULT 1,
    event_id uuid not null ,
    created_at timestamp without time zone not null default now(),
    UNIQUE (stream_id, stream_version),
    CONSTRAINT fk_event_stream_event_id FOREIGN KEY (event_id) REFERENCES event_store(event_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_stream;
-- +goose StatementEnd
