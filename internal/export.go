package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

type workflowBuild struct {
	ID             string    `json:"id"`
	Pipeline       string    `json:"pipelineName"`
	Started        time.Time `json:"started"`
	Finished       time.Time `json:"finished"`
	Status         string    `json:"status"`
	Sha            string    `json:"revision"`
	Trigger        string    `json:"trigger"`
	PullRequestURL string    `json:"pullRequestUrl"`
	SystemEvents   []struct {
		Kind    string `json:"kind"`
		Message string `json:"message"`
		Step    string `json:"step"`
	} `json:"systemEvents"`
	PendingApprovals []struct {
		Name    string `json:"name"`
		Title   string `json:"title"`
		Timeout struct {
			FinalState string `json:"finalState"`
		} `json:"timeout"`
	} `json:"pendingApprovals"`
}

type workflowResult struct {
	Workflows struct {
		Page   int             `json:"page"`
		Pages  int             `json:"pages"`
		Offset int             `json:"offset"`
		Limit  int             `json:"limit"`
		Total  int             `json:"total"`
		Builds []workflowBuild `json:"docs"`
	} `json:"workflows"`
}

func (b workflowBuild) ToModel(refType string, customerID string, integrationInstanceID string) (*sdk.CICDBuild, error) {
	var build sdk.CICDBuild
	build.ID = sdk.NewCICDBuildID(customerID, refType, b.ID)
	build.CustomerID = customerID
	build.IntegrationInstanceID = sdk.StringPointer(integrationInstanceID)
	build.RefType = refType
	build.RefID = b.ID
	build.URL = sdk.StringPointer("https://g.codefresh.io/build/" + b.ID)
	build.Sha = b.Sha
	build.Automated = b.Trigger == "build"
	build.PullrequestURL = sdk.StringPointer(b.PullRequestURL)
	sdk.ConvertTimeToDateModel(b.Started, &build.StartDate)
	if !b.Finished.IsZero() {
		sdk.ConvertTimeToDateModel(b.Finished, &build.EndDate)
	}
	switch b.Status {
	case "success":
		build.Status = sdk.CICDBuildStatusPass
	case "error":
		build.Status = sdk.CICDBuildStatusFail
		if len(b.SystemEvents) > 0 {
			msg := make([]string, 0)
			for _, e := range b.SystemEvents {
				if e.Kind == "error" {
					msg = append(msg, fmt.Sprintf("%s: %s", e.Step, e.Message))
				}
			}
			build.Message = sdk.StringPointer(strings.Join(msg, ","))
		}
	case "terminated":
		build.Status = sdk.CICDBuildStatusCancel
		if len(b.PendingApprovals) > 0 {
			msg := make([]string, 0)
			// see if it was terminated because the approval step was denied
			for _, e := range b.PendingApprovals {
				if e.Timeout.FinalState == "denied" {
					msg = append(msg, fmt.Sprintf("%s: Approval Denied", e.Name))
				}
			}
			build.Message = sdk.StringPointer(strings.Join(msg, ","))
		}
	}
	return &build, nil
}
