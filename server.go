package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

// serverTmpl stores templates for the web interface.
var serverTmpl = new(template.Template)

func init() {
	serverTmpl.Funcs(template.FuncMap{
		"format": func(num int64, base int) string {
			return strconv.FormatInt(num, base)
		},

		"shortlen": func() string {
			return flags.short.String()
		},
	})

	template.Must(serverTmpl.New("main").Parse(`<html>
	<head>
		<title>{{.Title}} :: Main</title>

		<script type='application/javascript' src='https://ajax.googleapis.com/ajax/libs/jquery/2.2.2/jquery.min.js' defer></script>
		<script type='application/javascript' src='/ps2avglogin.js' defer></script>

		<style type='text/css'>
			hr
			{
				width:80%;
			}
		</style>
	</head>
	<body style='background-color:#EEEEEE;'>
	<div style='max-width:640px;margin-left:auto;margin-right:auto;'>
			<div id='loading'>
				<h2>Loading...</h2>
			</div>
			<div id='main' style='display:none;'>
				<div id='noshort'>
					<h1>Excluding short sessions:</h1>
					<h2>Average session: <span class='average'></span></h2>
					<h3>Calculated from <span class='num'></span> sessions.</h3>
					A session is short if it lasts less than {{shortlen}}.
				</div>

				<hr />

				<div id='total'>
					<h1>Including short sessions:</h1>
					<h2>Average session: <span class='average'></span></h2>
					<h3>Calculated from <span class='num'></span> sessions.</h3>
				</div>

				<hr />

				Currently tracking <span id='online'></span> active sessions.<br />
				Tracker runtime: <span id='runtime'></span>
			</div>
		</div>
	</body>
</html>`))
}

// logHandler returns an http.Handler that logs every request that
// gets sent to h.
func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Got %q request for %q", req.Method, req.URL)

		h.ServeHTTP(rw, req)
	})
}

// tmplHandler returns a handler that serves the template t in
// serverTmpl.
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

// serveSession serves the current session as JSON.
func serveSession(rw http.ResponseWriter, req *http.Request) {
	e := json.NewEncoder(rw)
	err := e.Encode(<-session)
	if err != nil {
		log.Printf("Failed to write session: %v", err)
	}
}

// serveJS serves the javascript for the web interface.
func serveJS(rw http.ResponseWriter, req *http.Request) {
	_, err := io.WriteString(rw, `$(document).ready(function() {
	var loading = $('#loading');
	var main = $('#main');

	var noshort = {
		"average": $('#noshort .average'),
		"num": $('#noshort .num'),
	};
	var total = {
		"average": $('#total .average'),
		"num": $('#total .num'),
	};

	var online = $('#online');
	var runtime = $('#runtime');

	function getSession() {
		$.getJSON('/session', function(data) {
			loading.hide();
			main.show();

			noshort.average.html(data.noshort.cur);
			noshort.num.html(data.noshort.num);
			total.average.html(data.total.cur);
			total.num.html(data.total.num);

			online.html(data.numchars);
			runtime.html(data.runtime);
		});
	};

	getSession();
	setInterval(getSession, 5000);
});`)
	if err != nil {
		log.Printf("Failed to write JS: %v", err)
	}
}

// server runs the web interface.
func server() {
	http.Handle("/session", logHandler(http.HandlerFunc(serveSession)))
	http.Handle("/ps2avglogin.js", logHandler(http.HandlerFunc(serveJS)))
	http.Handle("/", logHandler(tmplHandler("main")))

	log.Printf("Starting server at %q...", flags.addr)
	err := http.ListenAndServe(flags.addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
