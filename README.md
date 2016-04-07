GLEXA
=====

Glexa is an Alexa Skills webservice written in Go. Integrate Glexa with your go web server to create/host your own
Alexa commands (skills).


Usage Example
-------------
VerifyRequest is the middleware that handles the authentication of Alexa requests.

```go
func alexaHandler(w http.ResponseWriter, r *http.Request) {
	glexa.VerifyRequest(handleAlexaCommands)(w, r)
}

log.Fatal(http.ListenAndServeTLS(":443", "mycert.pem", "mykey.pem", alexaHandler))
```

Responding to an Alexa Request
------------------------------

```go

var handleAlexaCommands http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	body, err := glexa.ParseBody(r.Body)
	if err != nil {
		log.Printf("error: could not parse alexa request body: %q\n", err)
		http.Error(w, "", http.StatusBadRequest)
	}

	response := glexa.NewResponse()
	response.Response.ShouldEndSession = true

	if b.Request.IsLaunch() {
		response.Tell("I did not understand your command. Please try again.")
	}

	if b.Request.IsSessionEnded() {
		response.Tell("Good bye")
	}

	// do something awesome with intent
	if b.Request.IsIntent() {
		response.Tell("You are awesome!")
	}

	// json encode response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}
```
