package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	lua "github.com/yuin/gopher-lua"
)

func save(db *sql.DB, L *lua.LState) int {
	e := L.ToTable(1)
	entity := e.RawGetString("saveable")
	data := map[string]interface{}{}
	e.ForEach(func(k lua.LValue, v lua.LValue) {
		if k.String() == "saveable" {
			return
		}
		switch v.Type() {
		case lua.LTNumber:
			data[k.String()] = float64(v.(lua.LNumber))
		case lua.LTBool:
			data[k.String()] = bool(v.(lua.LBool))
		case lua.LTString:
			data[k.String()] = string(v.(lua.LString))
		}
	})
	asJson, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed to marshal entity save data for entity %s\n", entity)
	}
	_, err = db.Exec("INSERT INTO gamedata (gameid, entity, data) VALUES (?, ?, ?) ON CONFLICT(gameid, entity) DO UPDATE SET data=?", "gameid:0001", entity, asJson, asJson)
	if err != nil {
		log.Fatalf("Failed upserting entity %s data to database: %s\n", entity, err)
	}
	L.Push(lua.LNumber(0))
	return 1
}

func worker() {
	db, err := sql.Open("sqlite3", "./saves.db")
	if err != nil {
		panic("failed to open saves database")
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS gamedata (id INTEGER PRIMARY KEY, gameid STRING, entity STRING, data STRING, UNIQUE(gameid, entity))")
	if err != nil {
		panic("failed to create gamedata table")
	}

	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("save", L.NewFunction(func(L *lua.LState) int {
		save(db, L)
		return 0
	}))

	if err := L.DoFile("game.lua"); err != nil {
		panic(err)
	}

	for i := 1; i <= 100; i++ {
		if err := L.CallByParam(lua.P{
			Fn: L.GetGlobal("tick"), NRet: 0, Protect: true,
		}, lua.LNumber(0.1)); err != nil {
			panic(err)
		}
	}

}

func main() {
	fmt.Println("Hello.")
	// TODO: demonstrate running multiple workers
	worker()
}
