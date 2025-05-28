-- name: CreateTrack :exec
INSERT INTO tracks (name, artist, genre, duration_in_seconds, year, explicit, bpm, key)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);
--

-- name: AddToWorking :exec
INSERT INTO working (name, artist, genre, duration_in_seconds, year, explicit, bpm, key)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);

-- name: RemoveFromWorking :exec
DELETE FROM working WHERE working.name = $1;

-- name: GetTrack :one
SELECT * FROM tracks WHERE tracks.name = $1 AND tracks.artist = $2;
--

-- name: GetAllTracks :many
SELECT * FROM tracks;

-- name: GetAllWorking :many
SELECT * FROM tracks;

-- name: DeleteTrack :exec
DELETE FROM tracks WHERE tracks.name = $1 AND tracks.artist = $2;

-- name: CleanupTracks :exec
DELETE FROM tracks WHERE tracks.key = '';

-- name: CleanupWorking :exec
DELETE FROM working;