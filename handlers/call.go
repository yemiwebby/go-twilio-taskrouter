package handlers

import (
	"encoding/xml"
	"fmt"
	"go-twilio-taskrouter/config"
	"go-twilio-taskrouter/setup"
	"net/http"
	"strings"
)

type VoiceResponse struct {
	XMLName xml.Name `xml:"Response"`
	Gather  struct {
		NumDigits string `xml:"numDigits,attr"`
		Action    string `xml:"action,attr"`
		Say       string `xml:"Say"`
	} `xml:"Gather"`
	Say string `xml:"Say"`
	Pause struct {
		Length string `xml:"length,attr"`
	} `xml:"Pause"`
	Hangup string `xml:"Hangup"`
}

func IncomingCall(w http.ResponseWriter, r *http.Request) {
	response := VoiceResponse{}
	response.Gather.NumDigits = "1"
	response.Gather.Action = config.GetEnv("HOST_URL") + "/enqueue"
	response.Gather.Say = "Welcome to our service. For Programmable SMS, press 1. For Programmable Voice, press 2."

	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(response)
}

func EnqueueCall(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	digits := strings.TrimSpace(r.FormValue("Digits"))
	product := ""
	workerName := ""

	client := setup.InitTwilioClient()
	workspaceSID, err := setup.GetWorkspaceSID(client, "Twilio Center Workspace")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`<Response><Message>Error retrieving workspace: %s</Message></Response>`, err.Error())))
		return
	}

	switch digits {
	case "1":
		product = "ProgrammableSMS"
	case "2":
		product = "ProgrammableVoice"
	default:
		response := `
		<Response>
			<Say>Invalid selection. Please call again and select a valid option. Goodbye.</Say>
			<Hangup/>
		</Response>`
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(response))
		return		
	}

	workerSID := setup.FindAvailableWorkerBySkill(client, workspaceSID, product)
	if workerSID != "" {
		workerName = setup.GetWorkerName(client, workspaceSID, workerSID)
	}

	if workerSID == "" {
		response := fmt.Sprintf(`
		<Response>
			<Say>We are sorry, no agents are currently available for %s. Redirecting you to voicemail.</Say>
			<Redirect>%s/voicemail</Redirect>
		</Response>`, product, config.GetEnv("HOST_URL"))
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(response))
		return
	}

	response := fmt.Sprintf(`
	<Response>
		<Say>Thank you for calling. Please hold while we connect you to an agent skilled in %s.</Say>
		<Pause length="3"/>
		<Say>The available agent is %s.</Say>
		<Pause length="2"/>
		<Say>Thanks for calling us today. Hope you have a nice day.</Say>
		<Hangup/>
	</Response>`, product, workerName)

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(response))
}