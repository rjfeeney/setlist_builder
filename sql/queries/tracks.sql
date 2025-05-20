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

-- name: GetTrack :one
SELECT * FROM tracks WHERE tracks.name = $1 AND tracks.artist = $2;
--

-- name: DeleteTrack :exec
DELETE FROM tracks WHERE tracks.name = $1 ANd tracks.artist = $2;

-- name: CleanupTracks :exec
DELETE FROM tracks WHERE tracks.key = '';