package main

import (
	"go-twilio-taskrouter/config"
	"go-twilio-taskrouter/handlers"
	"go-twilio-taskrouter/setup"
	"log"
	"net/http"
)


func main() {
	config.LoadConfig()

	client := setup.InitTwilioClient()
	setup.ConfigureWorkspace(client)

	mux := http.NewServeMux()
	mux.HandleFunc("/incoming", handlers.IncomingCall)
	mux.HandleFunc("/enqueue", handlers.EnqueueCall)
	mux.HandleFunc("/sms", handlers.UpdateWorkerStatus)
	mux.HandleFunc("/voicemail", handlers.RedirectToVoicemail)
	mux.HandleFunc("/voicemail-complete", handlers.VoicemailComplete)

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

