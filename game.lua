local tiny = require("tiny")

local flightSystem = tiny.processingSystem()
flightSystem.filter = tiny.requireAll("flying", "position", "speed")
function flightSystem:process(e, dt) 
	e.position = e.position + e.speed * dt
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

local spaceship = {
	flying = true,
	position = 0,
	speed = 0.7,
	saveable = "spaceship:0001",
}

local fasterShip = {
	flying = true,
	position = 0,
	speed = 5.0,
	saveable = "spaceship:0002",
}

local world = tiny.world(flightSystem, saveSystem, spaceship, fasterShip)

function tick(dt)
	world:update(dt)
end