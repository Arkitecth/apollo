CREATE TABLE IF NOT EXISTS playlists(
	id bigserial PRIMARY KEY NOT NULL, 
	created_at timestamp(0) with time zone NOT NULL DEFAULT now(), 
	name text NOT NULL, 
	user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
	version integer NOT NULL DEFAULT 1
); 

