CREATE TABLE IF NOT EXISTS songs (
	id bigserial PRIMARY KEY NOT NULL,  
	created_at timestamp(0) WITH time zone NOT NULL default now(), 
	name text NOT NULL, 
	song_url text NOT NULL, 
	artist text NOT NULL, 
	thumbnail text NOT NULL, 
	version integer NOT NULL DEFAULT 1
); 
