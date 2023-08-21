// ┌┌┐┬─┐o┌┐┐
// ││││─┤││││
// ┘ ┘┘ ┘┘┘└┘

package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/vmihailenco/msgpack/v5"
)

// Database connection object
var databaseConnection *pg.DB

// Root application context
var ctx context.Context

// Root application context cancelation function
var ctxCancel context.CancelFunc

// Global application configuration
var configuration *Configuration

// Web server
var fiberApp *fiber.App

// Current cowboy
var cowboy *Cowboy

func init() {
	var err error
	if configuration, err = LoadConfiguration(); err != nil {
		log.Panicln(err)
	}
	var options *pg.Options
	if options, err = pg.ParseURL(configuration.DatabaseConnectionString); err != nil {
		log.Panicln(err)
	}
	databaseConnection = pg.Connect(options)
	// Create context that will be used for the whole application
	ctx, ctxCancel = context.WithCancel(databaseConnection.Context())
	// Verify database connection
	if err := databaseConnection.Ping(ctx); err != nil {
		log.Panicln(err)
	}
	// Initialize rand
	rand.Seed(time.Now().UnixNano())
	// Start web server
	fiberApp = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	log.Println("Initialized successfully")
	go exit()
}

func main() {
	if configuration.OperationMode == Regular {
		var err error

		for cowboy == nil {
			cowboy, _, err = FindAndLockCowboy()
			switch err.(type) {
			case ErrorNoCowboys:
				time.Sleep(time.Second * 2)
			case error:
				log.Fatalf("Failed to acquire a cowboy: %v\n", err)
			}
		}
		go Fight()
		// Listen for shots
		fiberApp.Post("/shot", func(c *fiber.Ctx) error {
			shot := &Shot{}
			if err = msgpack.Unmarshal(c.Body(), shot); err != nil {
				log.Printf("Failed to decode message from %v\n", c.IP())
				return c.Status(fiber.StatusUnprocessableEntity).SendString("")
			}
			// Cowboy takes shot
			if err = cowboy.TakeShot(shot); err != nil {
				log.Printf("Failed to receive a shot: %v\n", err)
				return err
			}
			var response []byte
			if response, err = msgpack.Marshal(shot); err != nil {
				log.Panicf("Failed to marshal response shot: %v\n", err)
			}
			// Send response
			c.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
			return c.Status(fiber.StatusOK).Send(response)
		})
		fiberApp.Listen(fmt.Sprintf(":%d", configuration.ServerPort))
		for {
		}
	} else {
		DatabaseSeed()
	}
}

func exit() {
	// Schedule closing connection
	defer ctxCancel()
	// Wait for interruption
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT,
		syscall.SIGQUIT, syscall.SIGSTOP, os.Interrupt)
	sig := <-signalCh
	log.Printf("Exiting: %v\n", sig)
	// Exit from the program
	os.Exit(0)
}

func Fight() {
	log.Println("I am", cowboy)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !cowboy.IsAlive {
				log.Printf("I am dead %s", cowboy.GetLifeEmoji())
				var err error
				var tx *pg.Tx
				if tx, err = databaseConnection.Begin(); err != nil {
					log.Fatalln(err)
				}
				// Make sure to close transaction if something goes wrong.
				defer tx.Close()
				// Remove myself from cowboys list
				if _, err = databaseConnection.Model(cowboy).WherePK().Delete(); err != nil {
					log.Panicln("Failed to remove myself from the cowboys:", err)
				}
				time.Sleep(time.Second)
				return
			}
			target := &Cowboy{}
			var err error
			// Find alive cowboy
			if err = databaseConnection.Model(target).
				Where("cowboy.id != ?", cowboy.Id).
				ColumnExpr("cowboy.*").
				Join("JOIN leases l").
				JoinOn("l.expires_at > now()").
				JoinOn("l.cowboy_id = cowboy.id").
				JoinOn("cowboy.is_alive = true").
				OrderExpr("RANDOM()").
				Limit(1).Select(); err != nil {
				if err == pg.ErrNoRows {
					var count int
					if count, err = databaseConnection.Model((*Cowboy)(nil)).Count(); err != nil {
						log.Panicln(err)
					}
					if count == 1 {
						log.Println("I won!")
						return
					}
					log.Println("Waiting for duel to start...")
					target = nil
				} else {
					log.Panicln(err)
				}
			}
			if target != nil {
				// Find its lease
				lease := &Lease{}
				if err = databaseConnection.Model(lease).
					Where("cowboy_id = ?", target.Id).
					Limit(1).Select(); err != nil {
					log.Panicln(err)
				}
				// Making a shot
				shot := &Shot{
					Damage:    cowboy.Damage,
					Shooter:   cowboy,
					Target:    target,
					CreatedAt: time.Now(),
				}

				url := fmt.Sprintf("http://%s/shot", lease.LockedBy)
				var request *http.Request
				var requestBody []byte
				var response *http.Response
				var responseBody []byte
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				client := &http.Client{}

				if requestBody, err = msgpack.Marshal(shot); err != nil {
					log.Panicln("Failed to marshal the shot:", err)
				}

				if request, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody)); err != nil {
					log.Println("Error making a request with context:", err)
				}

				if response, err = client.Do(request); err != nil {
					log.Println("Error making a shot request:", err)
				} else {
					defer response.Body.Close()
					if responseBody, err = ioutil.ReadAll(response.Body); err != nil {
						log.Println("Error reading response:", err)
					}

					if err = msgpack.Unmarshal(responseBody, shot); err != nil {
						log.Println("Error unmarshaling response:", err)
					}
					if shot.IsLetal {
						log.Printf("I have killed %s %v\n", shot.Target.Name, shot.Target.GetLifeEmoji())
					}
				}
			}
			time.Sleep(time.Second)
		}
	}
}
