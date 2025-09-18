package activities

import (
	"time"

	"github.com/sstent/go-garth/garth/client"
	"github.com/sstent/go-garth/garth/types"
)

type ActivitiesClient interface {
	GetActivities(start, end time.Time) ([]types.Activity, error)
	Login(email, password string) error
	SaveSession(filename string) error
}

var newClient func(domain string) (ActivitiesClient, error) = func(domain string) (ActivitiesClient, error) {
	return client.NewClient(domain)
}

func SetClient(f func(domain string) (ActivitiesClient, error)) {
	newClient = f
}
