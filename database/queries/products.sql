-- name: UpsertProduct :exec
INSERT INTO products (name, label, category, uri, created_at) 
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(name) DO UPDATE SET 
  name = excluded.name,
  label = excluded.label,
  category = excluded.category,
  uri = excluded.uri, 
  updated_at = excluded.created_at;

-- name: GetProductById :one
SELECT id, name, label, category, uri, created_at
FROM products WHERE id = $1 LIMIT 1;

-- name: GetProductsByLabel :many
SELECT id, name, label, uri
FROM products 
WHERE label ILIKE $1 
ORDER BY label ASC NULLS LAST
LIMIT 100;

-- name: GetWatchedProducts :many
SELECT products.id, products.name, MIN(products.uri) AS uri
FROM products
INNER JOIN watch_lists ON watch_lists.product_id = products.id
GROUP BY products.id;

-- name: GetProductsWithNewReleases :many
SELECT 
  p.id AS product_id,
  p.label AS product_label, 
  jsonb_agg(
    jsonb_build_object(
      'release_label', pv.release_label,
      'version', pv.version,
      'version_release_date', pv.version_release_date,
      'version_release_link', pv.version_release_link
    )
  ) AS product_versions
FROM products p
JOIN LATERAL (
  SELECT pv.release_label, pv.version, pv.version_release_date, pv.version_release_link
  FROM product_versions pv
  WHERE pv.created_at = $1
  AND pv.product_id = p.id
  ORDER BY pv.version_release_date DESC NULLS LAST, pv.release_date DESC NULLS LAST
  LIMIT 3
) pv ON true
WHERE p.id = ANY($2::int[])
GROUP BY p.id
ORDER BY p.label ASC NULLS LAST;
