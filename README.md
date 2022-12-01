Hello.

I'm trying to make a proof-of-concept for my Go game server/worker project.

The basic idea is that the server hosts multiple games for the signed up
players. Each game is responsible for its logic, demonstrated with an ECS for
some reason, and is written in Lua.

Definition of done:

- [x] a "game host" or "worker" running a Lua game with an ECS
- [x] save and load functionality, exports to file/database/whatever
- [ ] get input from player
- [x] ~~show state to player~~ this is not not needed because we can read state directly from the persistence layer (e.g. sqlite in this case)
- [x] notify the outside world about events happening in the game

## Makefile

- `make dev`: run the server
- `make view`: opens a watcher on the db file so you can see the games' state (needs sqlite3 command)
