CREATE INDEX IF NOT EXISTS songs_artist_idx ON songs USING GIN (to_tsvector('simple', artist)); 

CREATE INDEX IF NOT EXISTS songs_name_idx ON songs USING GIN (to_tsvector('simple', name))

