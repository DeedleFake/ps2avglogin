package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var serverTmpl = new(template.Template)

func init() {
	serverTmpl.Funcs(template.FuncMap{
		"format": func(num int64, base int) string {
			return strconv.FormatInt(num, base)
		},
	})

	template.Must(serverTmpl.New("main").Parse(`<html>
	<head>
		<title>{{.Title}} :: Main</title>

		<script type='application/javascript' src='https://ajax.googleapis.com/ajax/libs/jquery/2.2.2/jquery.min.js' defer></script>
		<script type='application/javascript' src='/ps2avglogin.js' defer></script>
	</head>
	<body style='background-color:#EEEEEE;'>
		<div id='loading'>
			<h2>Loading...</h2>
		</div>
		<div id='main' style='display:none;'>
			<div id='noshort'>
				<h1>Excluding short sessions:</h1>
				<h2>Current average: <span class='average'></span></h2>
				<h3>Calculated from <span class='num'></span> logouts.</h3>
			</div>

			<hr />

			<div id='total'>
				<h1>Including short sessions:</h1>
				<h2>Current average: <span class='average'></span></h2>
				<h3>Calculated from <span class='num'></span> logouts.</h3>
			</div>
		</div>
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

func serveAverage(rw http.ResponseWriter, req *http.Request) {
	e := json.NewEncoder(rw)
	err := e.Encode(<-session)
	if err != nil {
		log.Printf("Failed to write average: %v", err)
	}
}

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

	function getAverage() {
		$.getJSON('/average', function(data) {
			loading.hide();
			main.show();

			noshort.average.html(data.noshort.cur);
			noshort.num.html(data.noshort.num);
			total.average.html(data.total.cur);
			total.num.html(data.total.num);
		});
	};

	getAverage();
	setInterval(getAverage, 5000);
});`)
	if err != nil {
		log.Printf("Failed to write JS: %v", err)
	}
}

func server() {
	http.Handle("/average", logHandler(http.HandlerFunc(serveAverage)))
	http.Handle("/ps2avglogin.js", logHandler(http.HandlerFunc(serveJS)))
	http.Handle("/", logHandler(tmplHandler("main")))

	log.Printf("Starting server at %q...", flags.session)
	err := http.ListenAndServe(flags.addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
