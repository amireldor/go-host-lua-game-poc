package main

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func save(L *lua.LState) int {
	e := L.ToTable(1)
	fmt.Println("SAVEABLE ID", e.RawGetString("saveable"))
	e.ForEach(func(k lua.LValue, v lua.LValue) {
		if k.String() == "saveable" {
			return
		}
		fmt.Println("save", k, v)
	})
	L.Push(lua.LNumber(0))
	return 1
}

func worker() {
	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("save", L.NewFunction(save))

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
	worker()
}
