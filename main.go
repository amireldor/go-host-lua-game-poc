package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	lua "github.com/yuin/gopher-lua"
)

func save(db *sql.DB, gameid string, L *lua.LState) int {
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
	_, err = db.Exec("INSERT INTO gamedata (gameid, entity, data) VALUES (?, ?, ?) ON CONFLICT(gameid, entity) DO UPDATE SET data=?", gameid, entity, asJson, asJson)
	if err != nil {
		log.Fatalf("Failed upserting entity %s data to database: %s\n", entity, err)
	}
	L.Push(lua.LNumber(0))
	return 1
}

func load(db *sql.DB, gameid string, L *lua.LState) {
	rows, err := db.Query("SELECT entity, data FROM gamedata WHERE gameid=?", gameid)
	if err != nil {
		log.Fatalf("Failed to load game data for game %s: %s\n", gameid, err)
	}
	defer rows.Close()

	for rows.Next() {
		var entity string
		var data string
		err = rows.Scan(&entity, &data)
		if err != nil {
			log.Fatalf("Failed to load entity for game %s: %s\n", gameid, err)
		}
		var fromJson = map[string]interface{}{}
		err = json.Unmarshal([]byte(data), &fromJson)
		if err != nil {
			log.Fatalf("Failed unmarshalling data for entity %s in game %s :%s", entity, gameid, err)
		}

		entityAsTable := L.NewTable()
		entityAsTable.RawSetString("saveable", lua.LString(entity))

		for k, v := range fromJson {
			//lint:ignore S1034 I don't need the type later
			switch v.(type) {
			case float64:
				entityAsTable.RawSetString(k, lua.LNumber(v.(float64)))
			case bool:
				entityAsTable.RawSetString(k, lua.LBool(v.(bool)))
			case string:
				entityAsTable.RawSetString(k, lua.LString(v.(string)))
			}

			if err := L.CallByParam(lua.P{
				Fn: L.GetGlobal("addEntity"), NRet: 0, Protect: true,
			}, entityAsTable); err != nil {
				log.Fatalf("Failed to call addEntity for entity %s in game %s: %s\n", entity, gameid, err)
			}
		}
	}
}

func worker(gameid string, wg *sync.WaitGroup) {
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
		save(db, gameid, L)
		return 0
	}))

	if err := L.DoFile("game.lua"); err != nil {
		panic(err)
	}

	load(db, gameid, L)

	for i := 1; i <= 100; i++ {
		if err := L.CallByParam(lua.P{
			Fn: L.GetGlobal("tick"), NRet: 0, Protect: true,
		}, lua.LNumber(1.0)); err != nil {
			panic(err)
		}
	}

	wg.Done()
}

func main() {
	fmt.Println("Hello.")
	// TODO: demonstrate running multiple workers
	wg := sync.WaitGroup{}
	wg.Add(2)
	go worker("gameid:0001", &wg)
	go worker("gameid:0002", &wg)
	wg.Wait()
}
