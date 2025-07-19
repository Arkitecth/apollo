CREATE TABLE IF NOT EXISTS playlist_songs(
	id bigserial PRIMARY KEY NOT NULL, 
	playlist_id bigserial references playlists(id) ON DELETE CASCADE, 
	song_id bigserial references songs(id) ON DELETE CASCADE, 
	user_id bigserial references users(id) ON DELETE CASCADE 
); 
