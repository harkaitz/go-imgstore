package main
import (
	"github.com/jmoiron/sqlx"
	"github.com/gin-gonic/gin"
	"github.com/harkaitz/go-imgstore"
	"github.com/pborman/getopt/v2"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
	"os"
	"net/http"
)

const help = `imgstore [-a FILE [-s SIZE]][-g ID][-w]`

func main() {
	
	var db    *sqlx.DB
	var err    error
	var id     int64
	var data []byte
	var format string
	var name   string
	
	dFlag := getopt.String('d', "/tmp/imgstore.db", "Database file.")
	aFlag := getopt.String('a', ""                , "Add image.")
	sFlag := getopt.String('s', "600x"            , "Set image size.")
	gFlag := getopt.Int64 ('g', 0                 , "Get image.")
	wFlag := getopt.Bool  ('w'                    , "Open web service.")
	
	getopt.SetUsage(func() { fmt.Println(help + "\n") })
	getopt.Parse()
	
	if *aFlag == "" && *gFlag == 0 && *wFlag == false {
		getopt.Usage()
		return
	}
	
	db, err = sqlx.Open("sqlite3", *dFlag)
	if err != nil {
		return
	}
	defer db.Close()
	
	err = imgstore.CreateTables(db)
	if err != nil {
		log.Print(err)
		return
	}
	
	if *aFlag != "" {
		id, err = imgstore.AddImageFile(db, *sFlag, *aFlag)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("%v\n", id)
	}
	
	if *gFlag != 0 {
		data, format, err = imgstore.GetImage(db, *gFlag)
		if err != nil {
			log.Print(err)
			return
		}
		name = fmt.Sprintf("/tmp/imgstore-%v.%v", *gFlag, format)
		err = os.WriteFile(name, data, 0600)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("%v\n", name)
	}
	
	if *wFlag {
		r := gin.Default()
		imgstore.AddRoute(db, r)
		log.Print("http://127.0.0.1:8080/imgstore/<ID>")
		http.ListenAndServe(":8080", r)
	}
}
