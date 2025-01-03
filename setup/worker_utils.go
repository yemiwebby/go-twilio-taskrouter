package setup

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/taskrouter/v1"
)

func GetWorkspaceSID(client *twilio.RestClient, friendlyName string) (string, error) {
	params := &openapi.ListWorkspaceParams{}
	workspaces, err := client.TaskrouterV1.ListWorkspace(params)
	if err != nil {
		return "", fmt.Errorf("failed to list workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if *ws.FriendlyName == friendlyName {
			return *ws.Sid, nil
		}
	}
	return "", fmt.Errorf("workspace '%s' not found", friendlyName)
}

func GetWorkerSID(client *twilio.RestClient, workspaceSID, phoneNumber string) (string, error) {
	params := &openapi.ListWorkerParams{}
	workers, err := client.TaskrouterV1.ListWorker(workspaceSID, params)
	if err != nil {
		return "", fmt.Errorf("failed to list workers: %w", err)
	}

	for _, worker := range workers {
		if worker.Attributes != nil {
			var attributes map[string]interface{}
			if err := json.Unmarshal([]byte(*worker.Attributes), &attributes); err == nil {
				if attributes["contact_uri"] == phoneNumber {
					return *worker.Sid, nil
				}
			}
		}
	}
	return "", fmt.Errorf("worker with phone number '%s' not found", phoneNumber)
}

func UpdateWorkerActivity(client *twilio.RestClient, workspaceSID, workerSID, activityName string) error {
	params := &openapi.ListActivityParams{}
	activities, err := client.TaskrouterV1.ListActivity(workspaceSID, params)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	var activitySID string
	for _, activity := range activities {
		if *activity.FriendlyName == activityName {
			activitySID = *activity.Sid
			break
		}
	}

	if activitySID == "" {
		return fmt.Errorf("activity '%s' not found", activityName)
	}

	_, err = client.TaskrouterV1.UpdateWorker(workspaceSID, workerSID, &openapi.UpdateWorkerParams{
		ActivitySid: &activitySID,
	})
	if err != nil {
		return fmt.Errorf("failed to update worker activity: %w", err)
	}

	fmt.Printf("Worker %s activity updated to %s\n", workerSID, activityName)
	return nil
}


func FindWorkerBySkill(client *twilio.RestClient, workspaceSID, skill string) string {
	params := &openapi.ListWorkerParams{}
	workers, err := client.TaskrouterV1.ListWorker(workspaceSID, params)
	if err != nil {
		log.Fatalf("Failed to list workers: %s", err.Error())
	}

	for _, worker := range workers {
		if worker.Attributes != nil {
			var attributes map[string]interface{}
			if err := json.Unmarshal([]byte(*worker.Attributes), &attributes); err == nil {
				if products, ok := attributes["products"].([]interface{}); ok {
					for _, p := range products {
						if p == skill {
							return *worker.Sid
						}
					}
				}
			}
		}
	}
	return ""
}


func FindAvailableWorkerBySkill(client *twilio.RestClient, workspaceSID, skill string) string {
	params := &openapi.ListWorkerParams{}
	workers, err := client.TaskrouterV1.ListWorker(workspaceSID, params)
	if err != nil {
		log.Fatalf("Failed to list workers: %s", err.Error())
	}

	for _, worker := range workers {
		if worker.Attributes == nil || worker.ActivitySid == nil {
			continue
		}

		var attributes map[string]interface{}
		if err := json.Unmarshal([]byte(*worker.Attributes), &attributes); err != nil {
			log.Printf("Failed to parse worker attributes: %s", err.Error())
			continue
		}

		if products, ok := attributes["products"].([]interface{}); ok {
			for _, product := range products {
				if product == skill && *worker.ActivityName == "Available" {
					return *worker.Sid
				}
			}
		}
	}

	return ""
}


func GetWorkerName(client *twilio.RestClient, workspaceSID, workerSID string) string {
	worker, err := client.TaskrouterV1.FetchWorker(workspaceSID, workerSID)
	if err != nil {
		log.Printf("Failed to fetch worker details: %s", err.Error())
		return "Unknown"
	}
	return *worker.FriendlyName
}
