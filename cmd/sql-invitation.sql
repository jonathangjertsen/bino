-- name: GetInvitation :one
SELECT id
FROM invitation
WHERE email = $1
  AND expires > NOW()
;

-- name: GetInvitations :many
SELECT *, home.name AS home_name
FROM invitation
LEFT JOIN home
  ON home.id = invitation.home
WHERE expires > NOW()
ORDER BY created DESC
;

-- name: InsertInvitation :exec
INSERT INTO invitation (
  id,
  email,
  expires,
  created
) VALUES (
  @id,
  @email,
  @expires,
  @created
);

-- name: DeleteInvitation :exec
DELETE FROM invitation
WHERE id = $1
;

-- name: DeleteInvitationByEmail :exec
DELETE FROM invitation
WHERE email = $1
;

-- name: DeleteExpiredInvitations :execresult
DELETE FROM invitation
WHERE expires < NOW()
;
