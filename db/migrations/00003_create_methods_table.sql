-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS methods (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS methods;
-- +goose StatementEnd
