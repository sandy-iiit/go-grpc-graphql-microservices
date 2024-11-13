package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/http"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_URL"`
	CatalogURL string `envconfig:"CATALOG_URL"`
	OrderURL   string `envconfig:"ORDER_URL"`
}

func main() {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}
	s, err := NewGraphQLServer(config.AccountURL, config.CatalogURL, config.OrderURL)
	if err != nil {
		log.Fatalf("Error initializing GraphQL server: %v", err)
	}
	http.Handle("/graphql", handler.New(s.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
