-- +goose Up
UPDATE recipes
SET info = jsonb_set(
    info - 'waterTemp',
    '{water_temp}',
    '"100°C"',
    true
)
WHERE info->>'water_temp' IS NULL
   OR info->>'water_temp' = '';

-- +goose Down
UPDATE recipes
SET info = info - 'water_temp'
WHERE info->>'water_temp' = '100°C';
