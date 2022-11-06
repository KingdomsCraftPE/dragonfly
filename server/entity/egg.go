package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct {
	transform
	age   int
	close bool

	owner world.Entity

	c *ProjectileTicker
}

// NewEgg ...
func NewEgg(pos mgl64.Vec3, owner world.Entity) *Egg {
	s := &Egg{c: ProjectileConfig{
		Gravity: 0.03,
		Drag:    0.01,
		Damage:  0,
		HitFunc: eggParticles,
	}.NewTicker(), owner: owner}
	s.transform = newTransform(s, pos)

	return s
}

// Type returns EggType.
func (egg *Egg) Type() world.EntityType {
	return EggType{}
}

// Tick ...
func (egg *Egg) Tick(w *world.World, current int64) {
	egg.c.Tick(egg, &egg.transform, w, current)
}

// eggParticles spawns 6 particle.EggSmash at a trace.Result.
func eggParticles(res trace.Result, w *world.World, _ world.Entity) {
	for i := 0; i < 6; i++ {
		w.AddParticle(res.Position(), particle.EggSmash{})
	}
	// TODO: Spawn chicken(egg) 12.5% of the time.
}

// New creates a egg with the position, velocity, yaw, and pitch provided. It doesn't spawn the egg,
// only returns it.
func (*Egg) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	egg := NewEgg(pos, owner)
	egg.vel = vel
	return egg
}

// Explode ...
func (egg *Egg) Explode(src mgl64.Vec3, force float64, _ block.ExplosionConfig) {
	egg.mu.Lock()
	egg.vel = egg.vel.Add(egg.pos.Sub(src).Normalize().Mul(force))
	egg.mu.Unlock()
}

// Owner ...
func (egg *Egg) Owner() world.Entity {
	egg.mu.Lock()
	defer egg.mu.Unlock()
	return egg.owner
}

// EggType is a world.EntityType implementation for Egg.
type EggType struct{}

func (EggType) EncodeEntity() string { return "minecraft:egg" }
func (EggType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (EggType) DecodeNBT(m map[string]any) world.Entity {
	egg := NewEgg(nbtconv.Vec3(m, "Pos"), nil)
	egg.vel = nbtconv.Vec3(m, "Motion")
	return egg
}

func (EggType) EncodeNBT(e world.Entity) map[string]any {
	egg := e.(*Egg)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(egg.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(egg.Velocity()),
	}
}
