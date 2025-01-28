-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coffees (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    region text,
    img text NOT NULL,
    description text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coffees;
-- +goose StatementEnd
