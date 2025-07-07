CREATE TABLE IF NOT EXISTS playlists(
	id bigserial PRIMARY KEY NOT NULL, 
	created_at timestamp(0) with time zone NOT NULL DEFAULT now(), 
	name text NOT NULL, 
	userID integer NOT NULL, 
	version integer NOT NULL
); 
