package Remux

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// The soul of remux, initialize a new app with a port
type Remux struct {
	Port string
}

// The heart of remux, containing utilities to perform action
type Engine struct {
	writer  http.ResponseWriter
	request *http.Request
	Vars    map[string]string
}

// Provides a text output to the browser
func (u Engine) Text(s string) {
	u.writer.Write([]byte(s))
}

// Provides a "json" output to the browser, just make sure to provide data for it to convert
func (u Engine) Json(raw any) error {
	var data, err = json.Marshal(raw)
	u.writer.Header().Set("Content-Type", "application/json")
	u.writer.Write(data)
	return err
}

// Redirects a handler to a given url passed into this function
func (u Engine) Redirect(url string) {
	http.Redirect(u.writer, u.request, url, http.StatusMovedPermanently)
}

// Serves a file to a given url
func (u Engine) File(url string, data any) {
	var render, _ = template.ParseFiles(url)
	render.Execute(u.writer, data)
}

// Allows only the given method to be passed
func (u Engine) Method(methods ...string) {
	for i, v := range methods {
		if u.request.Method != strings.ToUpper(v) && i == len(methods)-1 {
			u.writer.WriteHeader(405)
			return
		}
	}
}

func (u Engine) Body(str any) {
	var decoder = json.NewDecoder(u.request.Body)
	decoder.Decode(str)
}

// Basic handler to handle incomimg requests
func (r Remux) Handle(route string, handler func(e Engine)) {
	// /greet/{bob}
	var ogroute = route
	if route != "/" {
		var splitted = convert(route)
		route = "/" + splitted[1] + "/"
	}
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		if route != "/" {
			var str = remove(convert(ogroute), 0)
			var newstr = remove(convert(TrimSuffix(r.URL.Path, "/")), 0)
			var matched = match(str, newstr)
			// var decoder = json.NewDecoder(r.Body)
			// var t struct{}
			// decoder.Decode(&t)
			handler(Engine{w, r, matched})
		} else {
			// var decoder = json.NewDecoder(r.Body)
			// var t struct{}
			// decoder.Decode(&t)
			handler(Engine{w, r, nil})
		}
	})
}

func (r Remux) FileServer(url string, fileUrl string) {
	var fs = http.FileServer(http.Dir(fileUrl))
	if strings.HasSuffix(url, "/") {
		http.Handle(url, http.StripPrefix(url, fs))
	} else {
		http.Handle(url+"/", http.StripPrefix(url+"/", fs))
	}
}

// Start your app 🔥!
func (r Remux) Serve() {
	log.Fatal(http.ListenAndServe("localhost:"+r.Port, nil))
}

func convert(s string) []string {
	var splitted = strings.Split(s, "/")
	// for i, v := range splitted {
	// 	if v == "" {
	// 		splitted = remove(splitted, i)
	// 	}
	// }
	return splitted
}

func remove(arr []string, index int) []string {
	return append(arr[:index], arr[index+1:]...)
}

func match(arr []string, newarr []string) map[string]string {
	var matched = make(map[string]string)
	for i, v := range arr {
		if strings.Contains(v, "{") || strings.Contains(v, "}") {
			// name}
			var first = strings.Split(v, "{")
			var second = strings.Split(first[1], "}")
			v = second[0]
		} else if strings.Contains(newarr[i], "{") || strings.Contains(newarr[i], "}") {
			// name}
			var first = strings.Split(v, "{")
			var second = strings.Split(first[1], "}")
			v = second[0]
		}
		matched[v] = newarr[i]
	}
	return matched
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
