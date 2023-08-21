// ┬ ┐┌┐┐o┬  ┐─┐
// │ │ │ ││  └─┐
// ┘─┘ ┘ ┘┘─┘──┘

package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/goombaio/namegenerator"
)

func DatabaseSeed() {
	models := []interface{}{
		(*Cowboy)(nil),
		(*Lease)(nil),
		(*Shot)(nil),
	}

	log.Println("Creating tables")
	for _, model := range models {
		if err := databaseConnection.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			FKConstraints: true,
			IfNotExists:   true,
		}); err != nil {
			log.Panicln(err)
		}
	}

	if configuration.GenerateRandomCowboys > 0 {
		log.Println("Generating random cowboys")

		var err error
		nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
		for i := 0; i < int(configuration.GenerateRandomCowboys); i++ {
			cowboy := &Cowboy{
				Name:   nameGenerator.Generate(),
				Damage: uint32(rand.Intn(30) + 20),
				Health: int64(rand.Intn(100) + 100),
			}
			if _, err = databaseConnection.Model(cowboy).Insert(cowboy); err != nil {
				log.Panicln(err)
			}
			if err = databaseConnection.Model(cowboy).
				Column("id").
				Where("? = ?", pg.Ident("name"), cowboy.Name).
				Select(); err != nil {
				log.Panicln(err)
			}
			log.Printf("Added %s", cowboy)
		}
	}
}

func (c Cowboy) GetLifeEmoji() string {
	if c.IsAlive {
		return "🤠"
	} else {
		return "🪦"
	}
}
