CREATE TABLE IF NOT EXISTS playlistsongs(
	id bigserial PRIMARY KEY NOT NULL, 
	playlist_id bigserial references playlists(id) ON DELETE CASCADE, 
	song_id bigserial references songs(id)
); 
