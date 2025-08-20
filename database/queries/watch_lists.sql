-- name: CreateWatchList :one
INSERT INTO watch_lists (chat_id, product_id, created_at) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: DeleteWatchList :exec
DELETE FROM watch_lists 
WHERE chat_id = $1 
AND product_id = $2;

-- name: IsWatchListExists :one
SELECT EXISTS(
SELECT 1 FROM watch_lists 
WHERE chat_id = $1 AND product_id = $2
LIMIT 1);

-- name: GetWatchList :many
SELECT 
  p.name AS product_name, 
  p.label AS product_label
FROM watch_lists wl
JOIN products p ON wl.product_id = p.id
WHERE wl.chat_id = $1
ORDER BY p.label ASC NULLS LAST;

-- name: GetWatchListsWithProductVersions :many
SELECT 
  p.id AS product_id,
  p.label AS product_label,
  p.eol_url AS product_eol_url,
  json_agg(
    json_build_object(
      'release_label', pv.release_label,
      'version', pv.version,
      'version_release_date', pv.version_release_date,
      'version_release_link', pv.version_release_link
    )
  ) AS product_versions
FROM watch_lists wl
JOIN products p ON wl.product_id = p.id
LEFT JOIN LATERAL (
  SELECT release_label, version, version_release_date, version_release_link
  FROM product_versions
  WHERE product_id = p.id
  ORDER BY version_release_date DESC NULLS LAST, release_date DESC NULLS LAST
  LIMIT 1
) pv ON true
WHERE wl.chat_id = $1
GROUP BY p.id
ORDER BY p.label ASC NULLS LAST;

-- name: GetWatchListsGroupedByChat :many
SELECT
  chat_id,
  json_agg(DISTINCT product_id) AS product_ids
FROM watch_lists
GROUP BY chat_id;