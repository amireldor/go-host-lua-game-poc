local tiny = require("tiny")

local flightSystem = tiny.processingSystem()
flightSystem.filter = tiny.requireAll("flying", "position", "speed")
function flightSystem:process(e, dt) 
	local prev = e.position
	e.position = e.position + e.speed * dt
	if prev < 10000 and e.position >= 10000 then
		notify("Ship " .. e.saveable .. " have reached the moon!")
	end
end

local saveSystem = tiny.processingSystem()
saveSystem.filter = tiny.requireAll("saveable")
saveSystem.time = 0
function saveSystem:process(e, dt) 
	self.time = self.time + dt
	if math.floor(self.time) % 5 == 0 then
		save(e)
	end
end

local world = tiny.world(flightSystem, saveSystem)

function newGame()
	world:clearEntities()
	world:addEntity({
		flying = true,
		position = 0,
		speed = 8.4,
		saveable = "niceship001"
	})
	world:addEntity({
		flying = true,
		position = 0,
		speed = 1,
		saveable = "niceship002"
	})
end

function tick(dt)
	world:update(dt)
end

function addEntity(e)
	world:addEntity(e)
end