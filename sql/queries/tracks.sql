-- name: CreateTrack :exec
INSERT INTO tracks (name, artist, genre, duration_in_seconds, year, explicit, bpm, original_key)
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

-- name: AddOriginalKey :exec
UPDATE tracks
SET
    original_key = $1
WHERE name = $2 AND artist = $3;

-- name: GetTracksWithSingers :many
SELECT
    t.name,
    t.artist,
    t.genre,
    t.duration_in_seconds,
    t.year,
    t.explicit,
    t.bpm,
    t.original_key,
    s.singer,
    s.key AS singer_key
FROM
    tracks t
JOIN
    singers s ON t.name = s.song AND t.artist = s.artist;

-- name: AddSingerToWorking :exec
UPDATE working
SET 
    singer = $1,
    singer_key = $2
WHERE name = $3 AND artist = $4;

-- name: AddTrackToWorking :exec
INSERT INTO working (name, artist, genre, duration_in_seconds, year, explicit, bpm, original_key)
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

-- name: GetAllTables :many
SELECT table_name as tablename
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- name: SumDurationForSinger :many
SELECT
  s.singer,
  SUM(t.duration_in_seconds) AS total_duration
FROM
  singers s
JOIN
  tracks t
  ON s.song = t.name AND s.artist = t.artist
WHERE
  s.singer = ANY($1::text[])
GROUP BY
  s.singer
ORDER BY
  total_duration DESC;

-- name: GetTrackFromName :one
SELECT * FROM tracks WHERE name ILIKE $1;

-- name: GetSingerCombos :many
SELECT singer, key from singers WHERE song = $1 and artist = $2 AND singer = ANY($3::text[]);

-- name: CheckSingers :one
SELECT NOT EXISTS (
  SELECT 1 FROM singers WHERE song = $1 AND artist = $2
);

-- name: CheckKeys :many
SELECT name, artist FROM tracks WHERE original_key = '' OR original_key IS NULL;

-- name: CountSingers :one
SELECT COUNT(DISTINCT singer)
FROM singers;

-- name: GetSingers :many
SELECT singer FROM singers;


-- name: AddToSingers :exec
INSERT INTO singers (song, artist, singer, key)
VALUES (
    $1,
    $2,
    $3,
    $4
);

-- name: RemoveFromWorking :exec
DELETE FROM working WHERE working.name = $1;

-- name: GetTrack :one
SELECT * FROM tracks WHERE tracks.name = $1 AND tracks.artist = $2;

-- name: GetWorking :one
SELECT * FROM working WHERE working.name = $1;

-- name: GetAllTracks :many
SELECT * FROM tracks;

-- name: GetAllWorking :many
SELECT * FROM working;

-- name: DeleteTrack :exec
DELETE FROM tracks WHERE tracks.name = $1 AND tracks.artist = $2;

-- name: CleanTracks :exec
DELETE FROM tracks WHERE tracks.original_key = '';

-- name: CleanSingers :exec
DELETE FROM singers WHERE singer = '';

-- name: ClearWorking :exec
DELETE FROM working;