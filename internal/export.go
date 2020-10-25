package internal

import (
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

type workflowBuild struct {
	ID           string    `json:"id"`
	Pipeline     string    `json:"pipelineName"`
	Started      time.Time `json:"started"`
	Finished     time.Time `json:"finished"`
	Status       string    `json:"status"`
	Sha          string    `json:"revision"`
	Trigger      string    `json:"trigger"`
	SystemEvents []struct {
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
	sdk.ConvertTimeToDateModel(b.Started, &build.StartDate)
	if !b.Finished.IsZero() {
		sdk.ConvertTimeToDateModel(b.Finished, &build.EndDate)
	}
	switch b.Status {
	case "success":
		build.Status = sdk.CICDBuildStatusPass
	case "error":
		build.Status = sdk.CICDBuildStatusFail
		// if len(b.SystemEvents) > 0 {
		// 	msg := make([]string, 0)
		// 	for _, e := range b.SystemEvents {
		// 		if e.Kind == "error" {
		// 			msg = append(msg, fmt.Sprintf("%s: %s", e.Step, e.Message))
		// 		}
		// 	}
		// }
	case "terminated":
		build.Status = sdk.CICDBuildStatusCancel
	}
	return &build, nil
}

// "systemEvents": [
// 	{
// 		"retriable": false,
// 		"_id": "5f944bdc6426ff563430817c",
// 		"kind": "error",
// 		"message": "Failed to push image registry.gitlab.com/pinpt/event-api:master",
// 		"step": "Pushing to users registry"
// 	},
// 	{
// 		"retriable": false,
// 		"_id": "5f944bdceca6b5623f84c66c",
// 		"kind": "error",
// 		"message": "Failed to push image: pinpt/event-api:master to the registry",
// 		"step": "Building Docker Image"
// 	}
// ],

// "pendingApprovals": [
// 	{
// 		"tokens": {
// 			"cfApiKeyTokenName": "cfApiKey_5f935c1b85ce560cbfdb7ccd",
// 			"engineTokenName": "wf_5f935c1b85ce560cbfdb7ccd"
// 		},
// 		"timeout": {
// 			"duration": 1,
// 			"finalState": "denied"
// 		},
// 		"name": "approval_for_push",
// 		"startedAt": "2020-10-23T22:44:39.047Z",
// 		"historySegmentStarted": "2020-10-23T22:41:41.824Z",
// 		"historySegmentElectionDate": "2020-10-23T22:41:31.631Z",
// 		"title": "Deploy to Stable?",
// 		"finishedAt": "2020-10-23T23:44:46.598Z"
// 	}
// ],
