// ┌─┐┌─┐┐ ┬┬─┐┌─┐┐ ┬
// │  │ │││││─││ │└┌┘
// └─┘┘─┘└┴┘┘─┘┘─┘ ┘

package main

import (
	"fmt"
	"log"
	"time"
)

type Cowboy struct {
	tableName struct{} `pg:"cowboys"`

	Id        uint64    `pg:"id,pk"`
	Name      string    `pg:"name,unique,notnull"`
	Health    int64     `pg:"health,notnull,use_zero"`
	Damage    uint32    `pg:"damage,notnull,use_zero"`
	IsAlive   bool      `pg:"is_alive,default:true,use_zero"`
	CreatedAt time.Time `pg:"default:now()"`
}

type ErrorCowboyIsDead struct {
	Target *Cowboy
}

func (e ErrorCowboyIsDead) Error() string {
	return fmt.Sprintf("%s is already dead!", e.Target.Name)
}

func (c Cowboy) String() string {
	return fmt.Sprintf("%s %s (%d) ❤️ %d 💥 %d", c.GetLifeEmoji(), c.Name, c.Id, c.Health, c.Damage)
}

func (c *Cowboy) TakeShot(shot *Shot) error {
	if !c.IsAlive {
		return ErrorCowboyIsDead{Target: c}
	}

	// Save old health
	shot.OldTargetHealth = c.Health
	// Check was it letal
	shot.IsLetal = c.Health <= int64(shot.Damage)
	// Subtract health
	c.Health -= int64(shot.Damage)
	if c.Health <= 0 {
		c.Health = 0
		c.IsAlive = false
	}
	// Update the database
	if _, err := databaseConnection.Model(c).WherePK().Update(); err != nil {
		return err
	}
	if _, err := databaseConnection.Model(shot).Insert(); err != nil {
		return err
	}
	// Update the target
	shot.Target = c
	log.Println(shot)
	return nil
}
