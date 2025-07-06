-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coffees (
   id bigserial PRIMARY KEY,
   user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
   info jsonb,
   version int NOT NULL DEFAULT 1,
   created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coffees;
-- +goose StatementEnd
