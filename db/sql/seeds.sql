INSERT INTO users (name, email, password_hash, activated) VALUES
    ('Hunter', 'hunter@gmail.com', '$2a$12$swsVl00IqCna.Uq5Pssh9erv5sT9raLq.my2nZFFGxiXErVcPH9Hy', false);

INSERT INTO methods (name) VALUES
    ('Pour Over'),
    ('Hario Switch');

-- INSERT INTO recipes(user_id, coffee_id, method_id, info) VALUES
--     (
--      (SELECT id FROM users WHERE email = 'hunter@gmail.com'),
--      (SELECT id FROM coffees WHERE name = 'Milky Cake'),
--      (SELECT id FROM methods WHERE name = 'Hario Switch'),
--      json_build_object(
--         'grams_in', '20',
--         'grams_out', '320',
--         'phase', json_build_object(
--             '1', json_build_object(
--                  'open', 'true',
--                  'time', '45',
--                  'amount', '160'
--                  ),
--             '2', json_build_object(
--                  'open', 'false',
--                  'time', '75',
--                  'amount', '160'
--                  ),
--             '3', json_build_object(
--                  'open', 'true',
--                  'time', '60',
--                  'amount', '0'
--                  )
--         )
--      )
--     );