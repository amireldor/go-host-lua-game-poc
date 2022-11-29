Hello.

I'm trying to make a proof-of-concept for my Go game server/worker project.

The basic idea is that the server hosts multiple games for the signed up
players. Each game is responsible for its logic, demonstrated with an ECS for
some reason, and is written in Lua.

Definition of done:

- [x] a "game host" or "worker" running a Lua game with an ECS
- [ ] save functionality, exports to file/database/whatever
- [ ] get input from player
- [ ] show state to player
