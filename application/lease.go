// ┬  ┬─┐┬─┐┐─┐┬─┐
// │  ├─ │─┤└─┐├─
// ┘─┘┴─┘┘ ┘──┘┴─┘

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Lease struct {
	tableName struct{} `pg:"leases"`

	Id        uint64    `pg:"id,pk"`
	CowboyId  uint64    `pg:"cowboy_id,on_delete:CASCADE,on_update:CASCADE"`
	Cowboy    *Cowboy   `pg:"fk:cowboy_id,notnull,rel:has-one"`
	ExpiresAt time.Time `pg:"expires_at,notnull,default:now()"`
	LockedBy  string    `pg:"locked_by,notnull"`
}

type ErrorNoCowboys struct {
}

func (e ErrorNoCowboys) Error() string {
	return "No cowboys available, waiting..."
}

func (l Lease) Watch() {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			l.ExpiresAt = time.Now().Add(time.Second * 7)
			if _, err = databaseConnection.Model(&l).WherePK().Update(); err != nil {
				log.Panicf("Could not refresh the lease: %v\n", err)
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func FindAndLockCowboy() (*Cowboy, *Lease, error) {
	var err error
	var tx *pg.Tx
	if tx, err = databaseConnection.Begin(); err != nil {
		log.Fatalln(err)
	}
	// Make sure to close transaction if something goes wrong.
	defer tx.Close()
	// Sleep a bit to avoid collisions
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))

	// Find cowboy without worker
	cowboy := &Cowboy{}
	if err = databaseConnection.Model(cowboy).
		Join("LEFT JOIN leases l").
		JoinOn("cowboy.id = l.cowboy_id").
		Where("l.expires_at < now()").
		WhereOr("l.expires_at is null").
		Limit(1).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil, ErrorNoCowboys{}
		}
	}
	log.Println("Selected a cowboy")

	// Acquire a lease
	lease := &Lease{
		CowboyId:  cowboy.Id,
		ExpiresAt: time.Now().Add(time.Second * 5),
		LockedBy:  fmt.Sprintf("%s:%d", configuration.WorkerName, configuration.ServerPort),
	}
	var updated orm.Result
	if updated, err = databaseConnection.Model(lease).Where("cowboy_id = ?", lease.CowboyId).Update(); err != nil {
		log.Panicln(err)
	} else if updated.RowsAffected() == 0 {
		if _, err = databaseConnection.Model(lease).Insert(); err != nil {
			log.Panicf("Could not acquire a lease: %v\n", err)
		}
	}
	log.Println("Acquired a lease")
	// Refresh lease object to get valid id
	if err = databaseConnection.Model(lease).Where("cowboy_id = ?", lease.CowboyId).Select(); err != nil {
		log.Panicln(err)
	}
	go lease.Watch()

	return cowboy, lease, nil
}
