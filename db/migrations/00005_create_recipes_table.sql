-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recipes (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    coffee_id bigint NOT NULL REFERENCES coffees ON DELETE CASCADE,
    method_id bigint NOT NULL REFERENCES methods ON DELETE CASCADE,
    info jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
-- +goose StatementEnd
