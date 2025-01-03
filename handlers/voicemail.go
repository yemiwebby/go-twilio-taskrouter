package handlers

import "net/http"

func RedirectToVoicemail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	response := `
		<Response>
			<Say>We are sorry. All our agents are currently busy. Please leave a brief message after the beep. Your message will end automatically after 10 seconds, or you can press the pound key to finish.</Say>
			<Record maxLength="10" finishOnKey="#" action="/voicemail-complete" />
			<Say>Thank you for your message. Goodbye.</Say>
			<Hangup/>
		</Response>
	`
	w.Write([]byte(response))
}

func VoicemailComplete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	response := `
		<Response>
			<Say>Thank you for your message. Goodbye.</Say>
			<Hangup/>
		</Response>
	`
	w.Write([]byte(response))
}