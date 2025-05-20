-- +goose Up
CREATE TABLE tracks (
    name TEXT NOT NULL,
    artist TEXT NOT NULL,
    genre TEXT[],
    duration_in_seconds INT NOT NULL,
    year TEXT NOT NULL,
    explicit BOOL NOT NULL DEFAULT false,
    bpm INT NOT NULL,
    key TEXT NOT NULL,
    CONSTRAINT PK_name_artist PRIMARY KEY(name,artist),
    dnp BOOL NOT NULL DEFAULT false
);

-- +goose Down
DROP TABLE tracks;