package main
import (
    "html/template"
    "net/http"
    "regexp"
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
    "log"
)
    	
type Page struct {
	Title string
    Body  []byte
}
   	
func (p *Page) save() error {
	log.Printf("save! ZOMG.")
	db, err := sql.Open("mysql", "ross:ross@/go_schema")
    checkErr(err)
    defer db.Close()
    err = db.Ping()
	checkErr(err)
	
    stmt, err := db.Prepare("INSERT page set title=?, text=?")
    
    if err != nil {    	
    	return err
    }
    res, err := stmt.Exec(p.Title, p.Body)
    if err != nil {
    	return err
    }
    defer stmt.Close()
    lastId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
	return nil
}
func getTitles() (map[int]string, error) {
	db, err := sql.Open("mysql", "ross:ross@/go_schema")
	checkErr(err)
	defer db.Close()
	rows, err := db.Query("select id, title from page")
	if err != nil {
		log.Printf("no titles returned")
		return nil, err
	}
	defer rows.Close()
	titles := make(map[int]string)
	for rows.Next() {    
    	var name string
    	var id int
    	err = rows.Scan(&id, &name)
    	if err != nil {
    		checkErr(err)
    	}
    	titles[id] = name
	}
	return titles, err
}
func getPageText(title string) ([]byte, error){	 	
    var temp string
    db, err := sql.Open("mysql", "ross:ross@/go_schema")
    checkErr(err)
    defer db.Close()

    rows, err := db.Query("SELECT text from page where title=?", title)

    if err != nil {
    	log.Printf("nil result")
    	return nil, err
    }
    defer rows.Close()
    for rows.Next() {
    	//log.Printf("rows.")
    	err := rows.Scan(&temp)
    	if err != nil {
    		return nil, err
    	}
    	
    	log.Printf(temp)    	
    }
    return []byte(temp), err
}
func loadPage(title string) (*Page, error) {   	
    body, err := getPageText(title)
    if err != nil {
    	return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "list.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
    
var validPath = regexp.MustCompile("^/(edit|save|view|pages)/([a-zA-Z0-9]+)$")

func pagesHandler(w http.ResponseWriter, r *http.Request, title string) {
	log.Printf("pagesHandler")
	titles, err := getTitles()
	checkErr(err)
	err = templates.ExecuteTemplate(w, "list.html", titles)
    if err != nil {
    	checkErr(err)
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
    	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    	return
    }
    renderTemplate(w, "view", p)
}
   
func editHandler(w http.ResponseWriter, r *http.Request, title string) {   
    p, err := loadPage(title)
    if err != nil {
    	p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler (fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return 
		}
		fn(w, r, m[2])
	}
}
// for fatal errors
func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

/*func getDB() *sql.DB {
	
	db, err := sql.Open("mysql", "ross:ross@/go_schema")
    checkErr(err)
    defer db.Close()
    err = db.Ping()
	checkErr(err)
	
	return db
}*/
func main() {
	
	//handler stuffs
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/pages/", makeHandler(pagesHandler))
    http.ListenAndServe(":8080", nil)
}
 