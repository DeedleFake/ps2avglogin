package main

import (
	"log"
	"net/http"
	"text/template"
)

var serverTmpl = new(template.Template)

func init() {
	template.Must(serverTmpl.New("main").Parse(`<html>
	<head>
		<title>{{.Title}} :: Main</title>
	</head>
	<body>
		<h2>Unimplemented.</h2>
	</body>
</html>`))
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Got %q request for %q", req.Method, req.URL)

		h.ServeHTTP(rw, req)
	})
}

func tmplHandler(t string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := serverTmpl.ExecuteTemplate(rw, t, map[string]interface{}{
			"Req":   req,
			"Title": "PS2 Average Login Times",
		})
		if err != nil {
			log.Printf("Failed to execute %q: %v", t, err)
		}
	})
}

func server() {
	const (
		ServerAddr = ":8080"
	)

	http.Handle("/", logHandler(tmplHandler("main")))

	log.Printf("Starting server at %q...", ServerAddr)
	err := http.ListenAndServe(ServerAddr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
