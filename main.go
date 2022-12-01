package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

func load(db *sql.DB, gameid string, L *lua.LState) int {
	rows, err := db.Query("SELECT entity, data FROM gamedata WHERE gameid=?", gameid)
	if err != nil {
		log.Fatalf("Failed to load game data for game %s: %s\n", gameid, err)
	}
	defer rows.Close()

	entities := 0
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
			entities += 1
		}
	}
	return entities
}

func worker(gameid string, commands chan string) {
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

	L.SetGlobal("notify", L.NewFunction(func(L *lua.LState) int {
		m := L.ToString(1)
		log.Printf("[%s] %s\n", gameid, m)
		return 0
	}))

	if err := L.DoFile("game.lua"); err != nil {
		panic(err)
	}

	entitiesLoaded := load(db, gameid, L)
	if entitiesLoaded == 0 {
		if err := L.CallByParam(lua.P{
			Fn: L.GetGlobal("newGame"), NRet: 0, Protect: true,
		}); err != nil {
			panic(err)
		}
	}

	// sorry for non-idiomatic Go code
	run := true
	for run {
		select {
		case <-time.After(1 * time.Second):
			if err := L.CallByParam(lua.P{
				Fn: L.GetGlobal("tick"), NRet: 0, Protect: true,
			}, lua.LNumber(1.0)); err != nil {
				panic(err)
			}
		case cmd := <-commands:
			switch cmd {
			case "q":
				run = false
				commands <- "q" // let others consume the quit command too
			default:
				fmt.Printf("wooho! command for %s\n", gameid)
			}
		}
	}
}

func inputLoop(chs map[string]chan string) {
	// sorry for non-idiomatic Go code
	for {
		var cmd string
		var data string
		fmt.Print("(q to quit, c <gameid> for sending command): ")
		fmt.Scanln(&cmd, &data)
		switch cmd {
		case "q":
			for _, v := range chs {
				v <- "q"
			}
			fmt.Println("Bye.")
			return
		case "c":
			if ch, ok := chs[data]; ok {
				ch <- data
			}
		}
	}
}

func processGame(gameid string, chs map[string]chan string) {
	commands := make(chan string)
	go worker(gameid, commands)
	chs[gameid] = commands
}

func main() {
	fmt.Println("Hello.")
	chs := make(map[string]chan string)
	processGame("g1", chs)
	processGame("g2", chs)
	processGame("g3", chs)
	processGame("g4", chs)
	inputLoop(chs)
}
