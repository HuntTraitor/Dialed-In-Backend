INSERT INTO users (name, email, password_hash, activated) VALUES
    ('Hunter', 'hunter@gmail.com', '$2a$12$swsVl00IqCna.Uq5Pssh9erv5sT9raLq.my2nZFFGxiXErVcPH9Hy', false);

INSERT INTO methods (name, img) VALUES
    ('Pour Over', 'https://dialedin.s3.us-east-2.amazonaws.com/methods/pour_over.webp'),
    ('Hario Switch', 'https://dialedin.s3.us-east-2.amazonaws.com/methods/hario_switch.webp');

INSERT INTO coffees(user_id, name, region, process, img, description) VALUES
    ((SELECT id FROM users WHERE email = 'hunter@gmail.com'), 'Milky Cake', 'Columbia', 'Thermal Shock', 'https://res.cloudinary.com/dak-coffee-roasters/image/upload/f_auto,q_auto,c_scale,w_500//Products/Thumbs/limeball_nbz7sk', 'Very sweet coffee with notes of cinnamon');

INSERT INTO recipes(user_id, coffee_id, method_id, info) VALUES
    (
     (SELECT id FROM users WHERE email = 'hunter@gmail.com'),
     (SELECT id FROM coffees WHERE name = 'Milky Cake'),
     (SELECT id FROM methods WHERE name = 'Hario Switch'),
     json_build_object(
        'grams_in', '20',
        'grams_out', '320',
        'phase', json_build_object(
            '1', json_build_object(
                 'open', 'true',
                 'time', '45',
                 'amount', '160'
                 ),
            '2', json_build_object(
                 'open', 'false',
                 'time', '75',
                 'amount', '160'
                 ),
            '3', json_build_object(
                 'open', 'true',
                 'time', '60',
                 'amount', '0'
                 )
        )
     )
    );