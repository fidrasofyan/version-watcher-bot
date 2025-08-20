-- name: CreateProductVersion :exec
INSERT INTO product_versions (
  product_id, 
  release_name, 
  release_codename,
  release_label,
  release_date,
  version,
  version_release_date,
  version_release_link,
  created_at
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (version, release_name, product_id) DO NOTHING;

-- name: GetDistinctProductIdsFromProductVersionsByCreatedAt :many
SELECT DISTINCT product_id
FROM product_versions
WHERE created_at = $1;

