package ghlabels

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Labels is a type alias for a slice of github label pointers
type Labels []*github.Label

// GHLabeler is a wrapper around a github.Client used to interact with the
// Github label
type GHLabeler struct {
	Client *github.Client
	Config *Config
}

// NewLabeler returns a new GHLabeler instance
func NewLabeler(token string, c *Config) *GHLabeler {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	hostURL, err := url.Parse(c.Host)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	client.BaseURL = hostURL
	return &GHLabeler{Client: client, Config: c}
}

// GetLabels returns the list of existing labels from a specified gh repository
func (l *GHLabeler) GetLabels(repo string) Labels {
	labels, _, err := l.Client.Issues.ListLabels(l.Config.Owner, repo, nil)
	if err != nil {
		log.Printf("Error: Unable to Access %s/%s\n", l.Config.Owner, repo)
	}
	return labels
}

// SetLabels orchestrates setting the desired state of the provided repository's
// github issue labels
func (l *GHLabeler) SetLabels(existing Labels, repo string, removeAbsent bool) {
	l.CreateMissing(existing, repo)
	l.UpdateExisting(existing, repo)
	if removeAbsent {
		l.RemoveAbsent(existing, repo)
	}
}

// CreateMissing creates any github issue labels that don't already exist
func (l *GHLabeler) CreateMissing(existing Labels, repo string) {
	var create bool
	for _, label := range l.Config.Labels {
		create = true
		for _, el := range existing {
			if *label.Name == *el.Name {
				create = false
			}
		}
		if create {
			_, _, err := l.Client.Issues.CreateLabel(l.Config.Owner, repo, label)
			if err != nil {
				log.Printf("Error: Unable to Create Label for %s/%s\n", l.Config.Owner, repo)
			}
		}
	}
}

// UpdateExisting labels in github if a label color has been changed
func (l *GHLabeler) UpdateExisting(existing Labels, repo string) {
	for _, label := range l.Config.Labels {
		for _, el := range existing {
			if *label.Name == *el.Name && *label.Color != *el.Color {
				_, _, err := l.Client.Issues.EditLabel(l.Config.Owner, repo, *label.Name, label)
				if err != nil {
					log.Printf("Error: Unable to Access %s/%s\n", l.Config.Owner, repo)
				}
			}
		}
	}
}

// RemoveAbsent removes any labels that exist on github that aren't present in
// the local config file
func (l *GHLabeler) RemoveAbsent(existing Labels, repo string) {
	var delete bool

	for _, el := range existing {
		delete = true
		for _, label := range l.Config.Labels {
			if *label.Name == *el.Name {
				delete = false
			}
		}
		if delete {
			_, err := l.Client.Issues.DeleteLabel(l.Config.Owner, repo, *el.Name)
			if err != nil {
				log.Printf("Error: Unable to Access %s/%s\n", l.Config.Owner, repo)
			}
		}
	}
}
