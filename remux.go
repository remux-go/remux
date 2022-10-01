package remux

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	Query   url.Values
}

type Route struct {
	Url    string
	GET    func(e Engine)
	POST   func(e Engine)
	PUT    func(e Engine)
	DELETE func(e Engine)
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

// Get the post body of incoming (post) requests
func (u Engine) Body(str any) {
	var decoder = json.NewDecoder(u.request.Body)
	decoder.Decode(str)
}

var mux = http.NewServeMux()

var routes = []Route{}

// Handle incoming GET requests
func (r Remux) Get(route string, handler func(e Engine)) {
	route = strings.Split(route, "{")[0]
	if len(routes) == 0 {
		routes = append(routes, Route{route, handler, nil, nil, nil})
	} else {
		for i, v := range routes {
			if v.Url == route && i == len(routes)-1 {
				routes[i].GET = handler
				break
			} else if v.Url != route && i == len(routes)-1 {
				routes = append(routes, Route{route, handler, nil, nil, nil})
			}
		}
	}
}

// Handle incoming POST requests
func (r Remux) Post(route string, handler func(e Engine)) {
	// route = strings.TrimSuffix(strings.Split(route, "{")[0], "/")
	route = strings.Split(route, "{")[0]
	if len(routes) == 0 {
		routes = append(routes, Route{route, nil, handler, nil, nil})
	} else {
		for i, v := range routes {
			if v.Url == route {
				routes[i].POST = handler
				break
			} else if v.Url != route && i == len(routes)-1 {
				routes = append(routes, Route{route, nil, handler, nil, nil})
			}
		}
	}
}

// Handle incoming PUT requests
func (r Remux) Put(route string, handler func(e Engine)) {
	route = strings.Split(route, "{")[0]
	if len(routes) == 0 {
		routes = append(routes, Route{route, nil, nil, handler, nil})
	} else {
		for i, v := range routes {
			if v.Url == route {
				routes[i].PUT = handler
				break
			} else if v.Url != route && i == len(routes)-1 {
				routes = append(routes, Route{route, nil, nil, handler, nil})
			}
		}
	}
}

// Handle incoming DELETE requests
func (r Remux) Delete(route string, handler func(e Engine)) {
	route = strings.Split(route, "{")[0]
	if len(routes) == 0 {
		routes = append(routes, Route{route, nil, nil, nil, handler})
	} else {
		for i, v := range routes {
			if v.Url == route {
				routes[i].DELETE = handler
				break
			} else if v.Url != route && i == len(routes)-1 {
				routes = append(routes, Route{route, nil, nil, nil, handler})
			}
		}
	}
}

// serve files at a given path handler
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
	for _, v := range routes {
		spinup(v)
	}
	log.Fatal(http.ListenAndServe("localhost:"+r.Port, mux))
}

func convert(s string) []string {
	var splitted = strings.Split(s, "/")
	return splitted
}

func remove(arr []string, index int) []string {
	return append(arr[:index], arr[index+1:]...)
}

func match(arr []string, newarr []string) map[string]string {
	var matched = make(map[string]string, 100)
	if len(arr) == len(newarr) {
		for i, v := range arr {
			if strings.Contains(v, "{") || strings.Contains(v, "}") {
				var first = strings.Split(v, "{")
				var second = strings.Split(first[1], "}")
				v = second[0]
			} else if strings.Contains(newarr[i], "{") || strings.Contains(newarr[i], "}") {
				var first = strings.Split(v, "{")
				var second = strings.Split(first[1], "}")
				v = second[0]
			}
			matched[v] = newarr[i]
		}
	}
	return matched
}

func spinup(v Route) {
	var route = v.Url
	var ogroute = v.Url
	route = strings.Split(route, "{")[0]
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		var query = r.URL.Query()
		if route != "/" {
			var str = remove(convert(ogroute), 0)
			var newstr = remove(convert(strings.TrimSuffix(r.URL.Path, "/")), 0)
			var matched = match(str, newstr)
			w.WriteHeader(200)
			switch r.Method {
			case "GET":
				if v.GET != nil {
					v.GET(Engine{w, r, matched, query})
				}
			case "POST":
				if v.POST != nil {
					v.POST(Engine{w, r, matched, query})
				}
			case "PUT":
				if v.PUT != nil {
					v.PUT(Engine{w, r, matched, query})
				}
			case "DELETE":
				if v.DELETE != nil {
					v.DELETE(Engine{w, r, matched, query})
				}
			}
		} else {
			switch r.Method {
			case "GET":
				if v.GET != nil {
					v.GET(Engine{w, r, nil, query})
				}
			case "POST":
				if v.POST != nil {
					v.POST(Engine{w, r, nil, query})
				}
			case "PUT":
				if v.PUT != nil {
					v.PUT(Engine{w, r, nil, query})
				}
			case "DELETE":
				if v.DELETE != nil {
					v.DELETE(Engine{w, r, nil, query})
				}
			}
		}
	})
}
