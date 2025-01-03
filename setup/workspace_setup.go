package setup

import (
	"encoding/json"
	"fmt"
	"go-twilio-taskrouter/config"
	"log"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/taskrouter/v1"
)

func ConfigureWorkspace(client *twilio.RestClient) {
	workspace := findOrCreateWorkspace(client)

	workers := []struct {
		Name    string
		Product string
		Contact string
	}{
		{"Bob", "ProgrammableSMS", "+123456789"},
		{"TestUser", "ProgrammableVoice", "+447776736645"},
	}

	workerSIDs := make(map[string]string)
	for _, worker := range workers {
		attributes := fmt.Sprintf(`{"products": ["%s"], "contact_uri": "%s"}`, worker.Product, worker.Contact)
		workerSIDs[worker.Contact] = findOrCreateWorker(client, *workspace.Sid, worker.Name, attributes)
	}

	taskQueues := map[string]string{
		"ProgrammableSMS":  "products HAS 'ProgrammableSMS'",
		"ProgrammableVoice": "products HAS 'ProgrammableVoice'",
	}
	queueSIDs := findOrCreateTaskQueues(client, *workspace.Sid, taskQueues)

	findOrCreateWorkflow(client, *workspace.Sid, queueSIDs)

	fmt.Println("Workspace setup completed successfully.")
}

func findOrCreateWorkspace(client *twilio.RestClient) *openapi.TaskrouterV1Workspace {
	friendlyName := "Twilio Center Workspace"
	eventCallbackURL := config.GetEnv("HOST_URL") + "/callback/events"

	workspaces, err := client.TaskrouterV1.ListWorkspace(&openapi.ListWorkspaceParams{})
	if err != nil {
		log.Fatalf("Failed to list workspaces: %s", err.Error())
	}

	for _, ws := range workspaces {
		if *ws.FriendlyName == friendlyName {
			fmt.Printf("Found existing workspace: %s\n", *ws.Sid)
			return &ws
		}
	}

	workspace, err := client.TaskrouterV1.CreateWorkspace(&openapi.CreateWorkspaceParams{
		FriendlyName:    &friendlyName,
		EventCallbackUrl: &eventCallbackURL,
	})
	if err != nil {
		log.Fatalf("Failed to create workspace: %s", err.Error())
	}

	fmt.Printf("Created new workspace: %s\n", *workspace.Sid)
	return workspace
}

func findOrCreateWorker(client *twilio.RestClient, workspaceSID, name, attributes string) string {
	existingWorkers, err := client.TaskrouterV1.ListWorker(workspaceSID, &openapi.ListWorkerParams{})
	if err != nil {
		log.Fatalf("Failed to list workers: %s", err.Error())
	}

	for _, worker := range existingWorkers {
		if *worker.FriendlyName == name {
			fmt.Printf("Found existing worker '%s' with SID: %s\n", name, *worker.Sid)
			return *worker.Sid
		}
	}

	worker, err := client.TaskrouterV1.CreateWorker(workspaceSID, &openapi.CreateWorkerParams{
		FriendlyName: &name,
		Attributes:   &attributes,
	})
	if err != nil {
		log.Fatalf("Failed to create worker '%s': %s", name, err.Error())
	}
	fmt.Printf("Worker '%s' created with SID: %s\n", name, *worker.Sid)
	return *worker.Sid
}

func findOrCreateTaskQueues(client *twilio.RestClient, workspaceSID string, queues map[string]string) map[string]string {
	queueSIDs := make(map[string]string)

	existingQueues, err := client.TaskrouterV1.ListTaskQueue(workspaceSID, &openapi.ListTaskQueueParams{})
	if err != nil {
		log.Fatalf("Failed to list task queues: %s", err.Error())
	}

	for _, queue := range existingQueues {
		queueSIDs[*queue.FriendlyName] = *queue.Sid
	}

	for name, expression := range queues {
		if _, exists := queueSIDs[name]; exists {
			fmt.Printf("Found existing task queue: %s\n", name)
			continue
		}

		queue, err := client.TaskrouterV1.CreateTaskQueue(workspaceSID, &openapi.CreateTaskQueueParams{
			FriendlyName: &name,
			TargetWorkers: &expression,
		})
		if err != nil {
			log.Fatalf("Failed to create task queue: %s", err.Error())
		}
		fmt.Printf("Created new task queue: %s\n", name)
		queueSIDs[name] = *queue.Sid
	}

	return queueSIDs
}

func findOrCreateWorkflow(client *twilio.RestClient, workspaceSID string, queueSIDs map[string]string) {
	workflowName := "Tech Support Workflow"

	existingWorkflows, err := client.TaskrouterV1.ListWorkflow(workspaceSID, &openapi.ListWorkflowParams{})
	if err != nil {
		log.Fatalf("Failed to list workflows: %s", err.Error())
	}

	for _, workflow := range existingWorkflows {
		if *workflow.FriendlyName == workflowName {
			fmt.Printf("Found existing workflow: %s\n", *workflow.Sid)
			return
		}
	}


	workflowConfig := map[string]interface{}{
        "task_routing": map[string]interface{}{
            "filters": []map[string]interface{}{
                {
                    "expression": "selected_product=='ProgrammableSMS'",
                    "targets": []map[string]string{
                        {"queue": queueSIDs["ProgrammableSMS"]},
                    },
                },
                {
                    "expression": "selected_product=='ProgrammableVoice'",
                    "targets": []map[string]string{
                        {"queue": queueSIDs["ProgrammableVoice"]},
                    },
                },
            },
            "default_filter": map[string]interface{}{
                "queue": queueSIDs["ProgrammableVoice"],
                "fallback_url": config.GetEnv("HOST_URL") + "/voicemail",
            },
        },
    }

	configJSON, err := json.Marshal(workflowConfig)
	if err != nil {
		log.Fatalf("Failed to marshal workflow configuration: %s", err.Error())
	}

	workflow, err := client.TaskrouterV1.CreateWorkflow(workspaceSID, &openapi.CreateWorkflowParams{
		FriendlyName: &workflowName,
		Configuration: stringPtr(string(configJSON)),
	})
	if err != nil {
		log.Fatalf("Failed to create workflow: %s", err.Error())
	}

	fmt.Printf("Created new workflow: %s\n", *workflow.Sid)
}

func stringPtr(s string) *string {
	return &s
}
