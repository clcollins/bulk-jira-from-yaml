package jira

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/andygrunwald/go-jira"
	"github.com/clcollins/bulk-jira-from-yaml/pkg/config"
	"sigs.k8s.io/yaml"

	"github.com/k0kubun/pp"
)

var apiPath string = "/rest/api/3"

// createClient returns a *jiraClient with transport
// for the host specified in the application configuration
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

// whoAmI returns the *jira.User for the currently authenticated credentials
func whoAmI(client *jira.Client) (*jira.User, error) {
	user, response, err := client.User.GetSelf()
	if err != nil {
		printResponse(response)
		return user, err
	}

	return user, err
}

// getProjects returns a *jira.Project by the project ID
func getProjects(client *jira.Client, projectID string) (*jira.Project, error) {
	project, response, err := client.Project.Get(projectID)
	if err != nil {
		printResponse(response)
		return project, err
	}

	return project, err
}

// getIssueById returns an issue from the specified project
func getIssueById(client *jira.Client, project *jira.Project, issueID string) (*jira.Issue, error) {
	options := &jira.GetQueryOptions{}

	issue, response, err := client.Issue.Get(issueID, options)
	if err != nil {
		printResponse(response)
		return issue, err
	}

	return issue, err
}

// createIssue creates an issue from a spec
// requires Browse projects and Create issues project permissions
func createIssue(client *jira.Client, issueSpec *jira.Issue) (*jira.Issue, error) {
	issue, response, err := client.Issue.Create(issueSpec)

	if err != nil {
		printResponse(response)
		return issue, err
	}

	return issue, err
}

// printResponse converts the *jira.Response to bytes and logs to the terminal
func printResponse(response *jira.Response) error {

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Println(string(bytes))

	return err
}

// printIssueAsYaml prints a Jira issue to the terminal in YAML format
func printIssueAsYaml(issue *jira.Issue) error {
	out, err := yaml.Marshal(&issue)

	if err != nil {
		return err
	}

	fmt.Println("---")
	fmt.Println(string(out))

	return err
}

func Run(yamlFile string) error {

	issues, err := loadIssuesFromFile(yamlFile)

	if err != nil {
		return err
	}

	client, err := createClient()
	if err != nil {
		return err
	}

	user, err := whoAmI(client)
	if err != nil {
		return err
	}
	pp.Println(user)

	for _, issue := range issues {
		pp.Println(issue.SpecId)

		i := &jira.Issue{
			Fields: &jira.IssueFields{
				//	Creator:     user,
				Summary:     issue.Spec.Fields.Summary,
				Description: issue.Spec.Fields.Description,
				Project: jira.Project{
					Key: issue.Spec.Fields.Project.Key,
				},
			},
		}

		if (issue.Spec.Fields.Type == jira.IssueType{}) {
			i.Fields.Type = jira.IssueType{
				Name: "Story",
			}
		} else {
			i.Fields.Type = issue.Spec.Fields.Type
		}

		i.Fields.Labels = []string{
			"off-boarding",
		}

		if issue.Links != nil {
			for _, link := range issue.Links {
				l := &jira.IssueLink{
					Type: jira.IssueLinkType{
						Name: link.Type,
					},
					OutwardIssue: getIssueBySpecId(issues, link.LinksTo),
					InwardIssue:  getIssueBySpecId(issues, issue.SpecId),
				}

				i.Fields.IssueLinks = append(i.Fields.IssueLinks, l)

				// inward issues will be the Spec of this issue, but we
				// must have an outward issue to link to
				if l.OutwardIssue == nil {
					return errors.New("Unable to find target issue (linksTo) for link: %v not found")
				}
			}

		}

		var response *jira.Response
		i, response, err = client.Issue.Create(i)

		if err != nil {
			printResponse(response)
			pp.Println(i)
			return err
		}

	}

	return nil

}

// getIssueBySpecId returns the *jira.Issue from the issue
// list from the spec.Id
func getIssueBySpecId(issues []issueSpec, specId int) *jira.Issue {
	for _, issue := range issues {
		if issue.SpecId == specId {
			return issue.Spec
		}
	}

	return nil
}

type link struct {
	LinksTo int    `json:"linksTo"`
	Type    string `json:",inline"`
}

type issueSpec struct {
	SpecId int         `json:"spec_id"`
	Spec   *jira.Issue `json:",inline"`
	Links  []link      `json:"links"`
}

// loadIssuesFromFile takes a file represented as a string
// opens the file and reads it, returning an issue slice
func loadIssuesFromFile(file string) ([]issueSpec, error) {
	var issues []issueSpec

	filename, err := filepath.Abs(file)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(data, &issues)

	if err != nil {
		return nil, err
	}

	return issues, nil
}
