package jira

import (
	"log"

	"github.com/andygrunwald/go-jira"
	"github.com/clcollins/bulk-jira-from-yaml/pkg/config"
)

var apiPath string = "/rest/api/3"

func Run() error {
	transport := jira.BasicAuthTransport{
		Username: config.AppConfig.Username,
		Password: config.AppConfig.Token,
	}

	client, err := jira.NewClient(transport.Client(), config.AppConfig.Host)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println("Connecting to:", config.AppConfig.Host)

	user, _, err := client.User.GetSelf()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("User:", user)

	x, _, err := client.Project.Get("OHSS")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Project:", x)

	return nil
}
