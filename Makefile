dev: FORCE
	go run .

view: FORCE
	watch 'sqlite3 saves.db "SELECT * FROM gamedata"'

FORCE: ;