package main

import (
    "github.com/scottgreenup/scottgreenup.com/blog"
    "github.com/gorilla/mux"

    "bytes"
    "flag"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "sort"
    "strconv"
)

var port = flag.Int("port", 80, "The port for the webserver to run on.")

var templates = template.Must(template.ParseGlob("content/template/*"))

type Page struct {
    Title string
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string) error {
    err := templates.ExecuteTemplate(w, name, nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }

    return nil
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s - %s\n", r.RemoteAddr, r.URL.Path)

    err := renderTemplate(w, r, "about")
    if err != nil {
        log.Println(err.Error())
        http.Error(w, http.StatusText(500), 500)
        return
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s - %s\n", r.RemoteAddr, r.URL.Path)

    err := renderTemplate(w, r, "index")
    if err != nil {
        log.Println(err.Error())
        http.Error(w, http.StatusText(500), 500)
        return
    }
}

type ByTimestamp []blog.MetaData
func (b ByTimestamp) Len() int { return len(b) }
func (b ByTimestamp) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByTimestamp) Less(i, j int) bool { return b[i].Timestamp > b[j].Timestamp }

func blogHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s - %s\n", r.RemoteAddr, r.URL.Path)

    // Get all the posts out of the directory
    files, _ := ioutil.ReadDir("./content/posts");
    posts := make(map[blog.MetaData][]string)
    var meta_data []blog.MetaData
    for _, f := range files {
        markup, meta, err := blog.ParseHTMLFromFile("./content/posts/" + f.Name())
        if err != nil {
            log.Printf("Error from ParseHTMLFromFile(): %+v", err);
        }
        posts[meta] = markup;
        meta_data = append(meta_data, meta)
    }

    // Print them to a buffer, inserting HTML appropriately
    sort.Sort(ByTimestamp(meta_data))
    var buf bytes.Buffer
    buf.WriteString("{{define \"blog_content\"}}")
    for k, v := range meta_data {

        // Write the post to the buffer, insert the timestamp after header
        buf.WriteString("<article>")
        markup := posts[v];
        for i := 0; i < len(markup); i++ {
            buf.WriteString(markup[i])
        }
        buf.WriteString("</article>")

        // Insert a divider between posts
        if k != len(meta_data)-1 {
            buf.WriteString("<hr />");
        }
    }
    buf.WriteString("{{end}}")
    templates.Parse(buf.String());

    if r.Method != "GET" {
        return
    }

    if len(r.URL.RawQuery) == 0 {
        err := renderTemplate(w, r, "blog")
        if err != nil {
            log.Println(err.Error())
        }
        return
    }

    err := renderTemplate(w, r, "blog")
    if err != nil {
        log.Println(err.Error())
        http.Error(w, http.StatusText(500), 500)
        return
    }
}

func singleBlogHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]

    log.Printf("%s - %s\n", r.RemoteAddr, r.URL.Path)

    // Get the file with md at the end
    markup, _, err := blog.ParseHTMLFromFile("./content/posts/" + name + ".md")

    var buf bytes.Buffer
    buf.WriteString("{{define \"blog_content\"}}")
    // Write the post to the buffer, insert the timestamp after header
    buf.WriteString("<article>")
    for i := 0; i < len(markup); i++ {
        buf.WriteString(markup[i])
    }
    buf.WriteString("</article>")
    buf.WriteString("{{end}}")
    templates.Parse(buf.String());

    if r.Method != "GET" {
        return
    }

    if len(r.URL.RawQuery) == 0 {
        err := renderTemplate(w, r, "blog")
        if err != nil {
            log.Println(err.Error())
        }
        return
    }

    err = renderTemplate(w, r, "blog")
    if err != nil {
        log.Println(err.Error())
        http.Error(w, http.StatusText(500), 500)
        return
    }
}

func notFound(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s - %s\n", r.RemoteAddr, r.URL.Path)
    fmt.Fprintf(w, "404 Not Found");
}

func main() {
    flag.Parse()
    r := mux.NewRouter()

    // Handler for page URLs
    r.HandleFunc("/about",          aboutHandler)
    r.HandleFunc("/about/",         aboutHandler)
    r.HandleFunc("/blog",           blogHandler)
    r.HandleFunc("/blog/",          blogHandler)
    r.HandleFunc("/blog/{name}",    singleBlogHandler)
    r.HandleFunc("/",               indexHandler)

    r.NotFoundHandler = http.HandlerFunc(notFound);

    // Handler for static content (i.e. css, img, js)
    r.PathPrefix("/static/").Handler(
        http.StripPrefix(
            "/static",
            http.FileServer(http.Dir("content/static"))))

    // Listen and serve on `port`
    port_string := strconv.Itoa(*port)
    log.Printf("Listening on port %s\n", port_string)
    http.ListenAndServe(":" + port_string, r)
}
