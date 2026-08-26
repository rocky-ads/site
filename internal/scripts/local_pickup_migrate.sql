-- One-shot: add local_pickup flag to Parts categories and ads.
-- Usage (jump server):
--   psql "$DATABASE_URL" -f /workspace/internal/scripts/local_pickup_migrate.sql

BEGIN;

UPDATE categories
SET facets = '["condition", "local_pickup", "price"]'
WHERE name IN (
  'Car & Truck Parts',
  'Motorcycle Parts',
  'Bicycle Parts',
  'Agricultural Equipment Parts'
);

DELETE FROM ad_facets WHERE key = 'shipping';

INSERT INTO ad_facets (ad_id, key, num, text)
SELECT a.id, 'local_pickup', 1, NULL
FROM ads a
JOIN categories c ON c.id = a.category_id
WHERE c.name IN (
  'Car & Truck Parts',
  'Motorcycle Parts',
  'Bicycle Parts',
  'Agricultural Equipment Parts'
)
ON CONFLICT (ad_id, key) DO UPDATE
  SET num = EXCLUDED.num, text = EXCLUDED.text;

UPDATE ads a
SET vector_metadata = COALESCE(a.vector_metadata, '{}'::jsonb)
  || '{"local_pickup": 1}'::jsonb
FROM categories c
WHERE c.id = a.category_id
  AND c.name IN (
    'Car & Truck Parts',
    'Motorcycle Parts',
    'Bicycle Parts',
    'Agricultural Equipment Parts'
  );

COMMIT;
