-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coffees (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    region text NOT NULL,
    process text NOT NULL,
    img text NOT NULL,
    description text,
    version int NOT NULL DEFAULT 1
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coffees;
-- +goose StatementEnd
