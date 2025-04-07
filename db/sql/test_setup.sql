CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);

CREATE TABLE IF NOT EXISTS methods (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL
);

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

CREATE TABLE IF NOT EXISTS recipes (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    coffee_id bigint NOT NULL REFERENCES coffees ON DELETE CASCADE,
    method_id bigint NOT NULL REFERENCES methods ON DELETE CASCADE,
    info jsonb,
    version int NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

INSERT INTO users (name, email, password_hash, activated) VALUES
    ('Hunter', 'hunter@gmail.com', '$2a$12$swsVl00IqCna.Uq5Pssh9erv5sT9raLq.my2nZFFGxiXErVcPH9Hy', false);

INSERT INTO methods (name) VALUES
                                    ('Pour Over'),
                                    ('Hario Switch');

-- INSERT INTO recipes(user_id, coffee_id, method_id, info) VALUES
--     (
--         (SELECT id FROM users WHERE email = 'hunter@gmail.com'),
--         (SELECT id FROM coffees WHERE name = 'Milky Cake'),
--         (SELECT id FROM methods WHERE name = 'Hario Switch'),
--         json_build_object(
--                 'grams_in', '20',
--                 'grams_out', '320',
--                 'phase', json_build_object(
--                         '1', json_build_object(
--                         'open', 'true',
--                         'time', '45',
--                         'amount', '160'
--                              ),
--                         '2', json_build_object(
--                                 'open', 'false',
--                                 'time', '75',
--                                 'amount', '160'
--                              ),
--                         '3', json_build_object(
--                                 'open', 'true',
--                                 'time', '60',
--                                 'amount', '0'
--                              )
--                          )
--         )
--     );