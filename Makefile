dev: FORCE
	go run .

view: FORCE
	watch -n 1 'sqlite3 --readonly saves.db "SELECT * FROM gamedata"'

FORCE: ;