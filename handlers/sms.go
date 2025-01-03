package handlers

import (
	"fmt"
	"go-twilio-taskrouter/setup"
	"net/http"
	"strings"
)

func UpdateWorkerStatus(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	phoneNumber := strings.TrimSpace(r.FormValue("From"))
	body := strings.TrimSpace(strings.ToLower(r.FormValue("Body")))


	client := setup.InitTwilioClient()

	workspaceSID, err := setup.GetWorkspaceSID(client, "Twilio Center Workspace")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`<Response><Message>Error retrieving workspace: %s</Message></Response>`, err.Error())))
		return
	}

	workerSID, err := setup.GetWorkerSID(client, workspaceSID, phoneNumber)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`<Response><Message>Error finding worker: %s</Message></Response>`, err.Error())))
		return
	}

	message := "Unrecognized command. Reply with 'on' to become available or 'off' to go offline."
	if body == "on" {
		err = setup.UpdateWorkerActivity(client, workspaceSID, workerSID, "Available")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`<Response><Message>Error updating worker status: %s</Message></Response>`, err.Error())))
			return
		}
		message = fmt.Sprintf("Worker %s is now available.", phoneNumber)
	} else if body == "off" {
		err = setup.UpdateWorkerActivity(client, workspaceSID, workerSID, "Offline")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`<Response><Message>Error updating worker status: %s</Message></Response>`, err.Error())))
			return
		}
		message = fmt.Sprintf("Worker %s is now offline.", phoneNumber)
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(fmt.Sprintf(`<Response><Message>%s</Message></Response>`, message)))
}
