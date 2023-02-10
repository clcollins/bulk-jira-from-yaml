package jira

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/andygrunwald/go-jira"
	"github.com/clcollins/bulk-jira-from-yaml/pkg/config"
	"gopkg.in/yaml.v2"
)

var apiPath string = "/rest/api/3"

func createClient() (*jira.Client, error) {
	transport := jira.BasicAuthTransport{
		Username: config.AppConfig.Username,
		Password: config.AppConfig.Token,
	}

	client, err := jira.NewClient(transport.Client(), config.AppConfig.Host)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	log.Println("Connecting to:", config.AppConfig.Host)

	return client, err
}

func whoAmI(client *jira.Client) (*jira.User, error) {
	user, response, err := client.User.GetSelf()
	if err != nil {
		printResponse(response)
		return user, err
	}

	return user, err
}

func getProjects() error {
	client, err := createClient()

	if err != nil {
		return err
	}

	project, _, err := client.Project.Get("OHSS")
	if err != nil {
		return err
	}

	log.Println("Project:", project)

	return err
}

// createIssue creates an issue from a spec
// requires Browse projects and Create issues project permissions
func createIssue(client *jira.Client, issueSpec *jira.Issue) error {
	issue, response, err := client.Issue.Create(issueSpec)

	if err != nil {
		printResponse(response)
		return err
	}

	log.Println("Issue: ", issue)

	return err
}

func printResponse(response *jira.Response) error {

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Println(string(bytes))

	return err
}

func retrieveIssue(client *jira.Client, issueID string) error {
	options := &jira.GetQueryOptions{}

	issue, _, err := client.Issue.Get(issueID, options)

	if err != nil {
		return err
	}

	fmt.Println(issue)

	out, err := yaml.Marshal(&issue)

	if err != nil {
		return err
	}

	fmt.Println("---")
	fmt.Println(string(out))

	return err

}

func Run() error {
	client, err := createClient()

	if err != nil {
		log.Fatal(err)
		return err
	}

	user, err := whoAmI(client)
	if err != nil {
		return err
	}

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee:    user,
			Description: "Test Issue 4",
			Type: jira.IssueType{
				Name: "Story",
			},
			Project: jira.Project{
				Key: "OHSS",
			},
			Summary: "This is a fourth test issue, maybe with an assignee",
		},
	}

	err = createIssue(client, &i)
	if err != nil {
		return err
	}

	return err

}

func RUN2(yamlFile string) error {
	filename, err := filepath.Abs(yamlFile)

	if err != nil {
		return err
	}

	yamlData, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	var issues []jira.Issue

	err = yaml.Unmarshal(yamlData, &issues)

	if err != nil {
		return err
	}

	log.Print("Issues: ", &issues)

	return err
}
