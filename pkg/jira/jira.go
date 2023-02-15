package jira

import (
	"errors"
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

// link represents a single "outward" *jira.IssueLink{}
// Outward links are from the current issue TO another issue
// eg current issue -> depends on -> other issue
type link struct {
	LinksTo int    `json:"linksTo"`
	Type    string `json:",inline"`
}

// issueSpec is a rough approximation of a jira issue, containing
// a *jira.Issue "spec", with some or all of the standard fields,
// a list of type link, and a SpecId which can be used to specify
// targets for links before a real Jira issue has been created and
// a real key (eg "PROJECT-##") has been created
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

// getProjects returns a *jira.Project by the project ID
func getProjects(client *jira.Client, projectID string) (*jira.Project, error) {
	project, response, err := client.Project.Get(projectID)
	if err != nil {
		printResponse(response)
		return project, err
	}

	return project, err
}

// getIssueBySpecId returns the *jira.Issue from the issue
// list from the SpecId of the issue type (from the parsed yaml)
// rather than the literal issue key
func getIssueBySpecId(issues []issueSpec, specId int) *jira.Issue {
	for _, issue := range issues {
		if issue.SpecId == specId {
			return issue.Spec
		}
	}

	return nil
}

// getIssueById returns an issue pointer from the specified project
func getIssueById(client *jira.Client, project *jira.Project, issueID string) (*jira.Issue, error) {
	options := &jira.GetQueryOptions{}

	issue, response, err := client.Issue.Get(issueID, options)
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

	pp.Println(string(bytes))

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

	for _, issue := range issues {
		pp.Println(issue.SpecId)

		i := &jira.Issue{
			Fields: &jira.IssueFields{
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
