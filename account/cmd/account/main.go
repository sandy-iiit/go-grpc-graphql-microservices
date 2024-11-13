package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
	"go-graphql-grpc-microservice/account"
	"log"
	"time"
)

type Config struct {
	DatabaseUrl string `envconfig:"DATABASE_URL"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(config.DatabaseUrl)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	defer r.Close()
	log.Println("Listening on :8080")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8080))

}
