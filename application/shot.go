// ┐─┐┬ ┬┌─┐┌┐┐
// └─┐│─┤│ │ │
// ──┘┘ ┴┘─┘ ┘

package main

import (
	"fmt"
	"time"
)

type Shot struct {
	tableName struct{} `pg:"shots"`

	Id              uint64    `pg:"id,pk"`
	ShooterId       uint64    `pg:"shooter_id,on_delete:CASCADE,on_update:CASCADE"`
	Shooter         *Cowboy   `pg:"fk:shooter_id,notnull,rel:has-one"`
	TargetId        uint64    `pg:"target_id,on_delete:CASCADE,on_update:CASCADE"`
	Target          *Cowboy   `pg:"fk:target_id,notnull,rel:has-one"`
	Damage          uint32    `pg:"damage,notnull,use_zero"`
	OldTargetHealth int64     `pg:"old_target_health,notnull,use_zero"`
	IsLetal         bool      `pg:"is_letal,default:false,use_zero"`
	CreatedAt       time.Time `pg:"default:now()"`
}

func (s Shot) String() string {
	shooterName := s.Shooter.Name
	targetName := s.Target.Name
	if shooterName == cowboy.Name {
		shooterName = "me"
	} else if targetName == cowboy.Name {
		targetName = "me"
	}
	return fmt.Sprintf("%s 💥 %s ❤️ %d (%d): %s",
		shooterName, targetName,
		s.Target.Health, s.Target.Health-s.OldTargetHealth,
		s.Target.GetLifeEmoji())
}
