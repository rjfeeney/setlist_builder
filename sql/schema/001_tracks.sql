-- +goose Up
CREATE TABLE tracks (
    name TEXT NOT NULL,
    artist TEXT NOT NULL,
    genre TEXT[],
    duration_in_seconds INT NOT NULL,
    year TEXT NOT NULL,
    explicit BOOL NOT NULL DEFAULT false,
    bpm INT NOT NULL DEFAULT 0,
    original_key TEXT NOT NULL DEFAULT '',
    CONSTRAINT PK_name_artist PRIMARY KEY(name,artist)
);

CREATE TABLE working (
    name TEXT NOT NULL,
    artist TEXT NOT NULL,
    genre TEXT[],
    duration_in_seconds INT NOT NULL,
    year TEXT NOT NULL,
    explicit BOOL NOT NULL DEFAULT false,
    bpm INT NOT NULL,
    original_key TEXT NOT NULL,
    singer TEXT,
    singer_key TEXT,
    CONSTRAINT PK_working PRIMARY KEY(name,artist)
);

CREATE TABLE singers (
    song TEXT NOT NULL,
    artist TEXT NOT NULL,
    singer TEXT NOT NULL,
    key TEXT NOT NULL,
    CONSTRAINT PK_singers PRIMARY KEY(song, artist, singer),
    CONSTRAINT FK_singers_tracks FOREIGN KEY (song, artist)
        REFERENCES tracks(name, artist)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE singers;
DROP TABLE working;
DROP TABLE tracks;