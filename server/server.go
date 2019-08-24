package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	cron_gql "github.com/henry74/cron-gql"
	"github.com/robfig/cron/v3"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	client := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	client.Start()
	emptyJobs := make(map[int]cron_gql.Job)

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(cron_gql.NewExecutableSchema(cron_gql.Config{Resolvers: &cron_gql.Resolver{Cron: client, RunningJobs: emptyJobs}})))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
