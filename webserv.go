package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_"github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
	"io/ioutil"
	"strings"
	"strconv"
)

type testStruct struct {
	Name    string `json:"Name"`
	Address string
}

type addStruct struct {
	Serie		string `json:"serie"`
	Ep			string `json:"ep"`
  Dir			string `json:"dir"`
	Subdir  string `json:"subdir"`
}

var db *sql.DB

func addHandler(w http.ResponseWriter, req *http.Request) {
	/*
	Legacy
	*/
	if req.Method == "GET" {
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err.Error())
	}
	log.Println(string(body))

	var addItem addStruct
	err = json.Unmarshal(body, &addItem)
	if err != nil {
		panic(err.Error())
	}
	log.Println(addItem.Ep + " " + addItem.Serie + " " + addItem.Dir)

	stmt, err := db.Prepare("INSERT INTO "+ addItem.Serie +" VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = stmt.Exec(addItem.Ep, addItem.Dir, addItem.Subdir)
	if err != nil {
		log.Fatal(err.Error())
	}

}

func serveSubtitle(w http.ResponseWriter, r *http.Request) {
	/*
	Serves the requested subtitle file.
	*/

	vars := mux.Vars(r)
	ep := vars["ep"]
	season := vars["s"]
	serie := vars["serie"]

	queryString := "SELECT sub_dir FROM "+ serie + " WHERE episode = "+ ep +" AND season = " + season

	subres, err := db.Query(queryString)

	if err != nil {
		panic(err.Error())
	}

	var subfile string

	for subres.Next() {
		err := subres.Scan(&subfile)
		if err != nil {
			panic(err.Error())
		}
	}

	if len(subfile) == 1 {
		w.WriteHeader(http.StatusNoContent)
	 	w.Write([]byte("☄ HTTP status code returned!"))
	}else {
		w.Header().Add("Content-Type", "TextTrack")
		http.ServeFile(w, r, subfile)
	}
}

func serveMedia(w http.ResponseWriter, r *http.Request) {
	/*
	Serves the requested media information.
	*/

	if r.Method == "GET" {
		panic("Wrong HTTP method")
	}

	type mediaInfo struct {
		VidUrl 		string
		SubUrl 		string
		SubLang		string
	}

	vars := mux.Vars(r)
	ep := vars["ep"]
	season := vars["s"]
	serie := vars["serie"]

	file := "/w/vid/" + serie + "/" + season + "/" + ep
	subfile := "/w/sub/" + serie + "/" + season + "/" + ep

	info := mediaInfo{file, subfile, "English"}

	mediaInfoJson, err := json.Marshal(info)
	if err != nil {
			panic(err.Error())
	}
	fmt.Fprintf(w, string(mediaInfoJson))
}

func serveVideo(w http.ResponseWriter, r *http.Request) {
	/*
	Serves the requsted video file.
	*/
	vars := mux.Vars(r)
	ep := vars["ep"]
	season := vars["s"]
	serie := vars["serie"]

	queryString := "SELECT media_dir FROM "+ serie + " WHERE episode = "+ ep +" AND season = " + season

	res, err := db.Query(queryString)

	if err != nil {
		log.Println("Query err")
		panic(err.Error())
	}
	var file string

	for res.Next() {
		err := res.Scan(&file)
		if err != nil {
			panic(err.Error())
		}
		//break // Some titles have more than one entry for some reason, the first one is the correct one
	}
	log.Println("File")
	log.Println(file)
	splitted := strings.Split(file, ".")
	w.Header().Add("Content-Type", "video/" + splitted[1] + "; codecs=ac3, a52, MPEG-4 AVC, avc1.4D401E, mp4a.40.2, H.264/AVC")

	http.ServeFile(w, r, file)
}

func getTitles(w http.ResponseWriter, r *http.Request) {
	/*
	Serves all the titles in the database.
	*/

	type title struct {
		Title 	string
		Seasons int
		Episodes_max []int
		Episodes_min []int
	}

	type titlesStruct struct {
		Titles	[]title
	}
	var titles titlesStruct
	res, err := db.Query("SELECT name FROM sqlite_master WHERE type = \"table\";")
	if err != nil {
		log.Println("title err")
		panic(err.Error())
	}
	titles.Titles = make([]title, 0)

	for res.Next() {
		var a title
		err = res.Scan(&a.Title)
		if err != nil {
			panic(err.Error())
		}

		seas, err := db.Query("SELECT MAX(season) FROM " + a.Title + ";")
		for seas.Next() {
			err = seas.Scan(&a.Seasons)
		}

		a.Episodes_max = make([]int, a.Seasons)
		a.Episodes_min = make([]int, a.Seasons)
		for i := 1; i <= a.Seasons; i++ {
			epi, err := db.Query("SELECT MAX(episode) FROM " + a.Title + " WHERE season = " + strconv.Itoa(i) + ";")
			if err != nil {
				panic(err.Error())
			}
			for epi.Next() {
				err = epi.Scan(&a.Episodes_max[i-1])
			}
			epi, err = db.Query("SELECT MIN(episode) FROM " + a.Title + " WHERE season = " + strconv.Itoa(i) + ";")
			if err != nil {
				panic(err.Error())
			}
			for epi.Next() {
				err = epi.Scan(&a.Episodes_min[i-1])
			}
		}

		if err != nil {
			log.Println("aaaa")
			panic(err.Error())
		}

		titles.Titles = append(titles.Titles, a)
	}
	titlesJson, _ := json.Marshal(titles)
	fmt.Fprintf(w, string(titlesJson))
}

func main() {

	db_temp, err := sql.Open("sqlite3", "media.db");
	db = db_temp

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/watch/{serie}/{s}/{ep}", serveMedia)
	r.HandleFunc("/w/vid/{serie}/{s}/{ep}", serveVideo)
	r.HandleFunc("/w/sub/{serie}/{s}/{ep}", serveSubtitle)
	r.HandleFunc("/add", addHandler)
	r.HandleFunc("/get_titles", getTitles)
	r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request){
		http.ServeFile(w, r, "html/index.html")
	})
	r.HandleFunc("/js/video.js", func (w http.ResponseWriter, r *http.Request){
		w.Header().Add("Content-Type", "application/javascript")
		http.ServeFile(w, r, "html/js/video.js")
	})
	r.HandleFunc("/styles.css", func (w http.ResponseWriter, r *http.Request){
		http.ServeFile(w, r, "html/styles.css")
	})

	http.Handle("/", r)


	http.ListenAndServe(":8080", nil)
}
