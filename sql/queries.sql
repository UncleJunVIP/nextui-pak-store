-- name: ListInstalledPaks :many
SELECT *
FROM installed_paks
WHERE can_uninstall = 1
ORDER BY name;

-- name: ListInstalledPaksWithoutRepo :many
SELECT *
FROM installed_paks
WHERE repo_url IS NULL;

-- name: Install :exec
INSERT INTO installed_paks (display_name, name, repo_url, version, type, can_uninstall)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateVersion :exec
UPDATE installed_paks
SET version = ?
WHERE repo_url = ?;

-- name: Uninstall :exec
DELETE
FROM installed_paks
WHERE repo_url = ?;

-- name: UpdateInstalledWithRepo :exec
UPDATE installed_paks
SET display_name = @new_display_name,
    name         = @new_name,
    repo_url     = @new_repo_url
WHERE display_name = @old_display_name;
