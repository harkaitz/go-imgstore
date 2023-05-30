package imgstore

import (
	"github.com/jmoiron/sqlx"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"os/exec"
	"bytes"
	"log"
	"os"
	"errors"
	"strconv"
	"database/sql"
	"io"
)

type Image struct {
	ID       int64   `db:"ID"        sql:"AUTO_INCREMENT" `
	Format   string  `db:"FORMAT"    sql:"NOT NULL"       `
	Data     []byte  `db:"DATA"      sql:"NOT NULL"       `
}

func CreateTables (db *sqlx.DB) (err error) {
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS images (
	    ID        INTEGER         PRIMARY KEY AUTOINCREMENT,
	    FORMAT    VARCHAR(16)     NOT NULL,
	    DATA      LONGBLOB        NOT NULL
	)`)
	return
}

var convertProg   string
var convertFormat string = "png"

func init() {
	var err error
	convertProg, err = exec.LookPath("convert")
	if err != nil {
		log.Fatal("ImageMagick is not installed.")
	}
}

func ConvertImageFile(size string, filename string, file io.Reader) (odata []byte, err error) {
	var cmd    *exec.Cmd
	var stdout  bytes.Buffer
	var stderr  bytes.Buffer
	var suffix  string
	
	suffix = filepath.Ext(filename)
	switch suffix {
	case ".jpg",".jpeg",".png":
	default: err = errors.New("Invalid format"); return
	}
	
	cmd = exec.Command(convertProg, "-resize", size, "-strip", suffix[1:] + ":-", convertFormat + ":-")
	cmd.Stdin  = file
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Start()
	if err != nil {
		return
	}
	
	err = cmd.Wait()
	if err != nil {
		log.Printf("convert: %s", stderr.String())
		return
	}
	
	odata = stdout.Bytes()
	if len(odata) > 512 * 1024 {
		err = errors.New("Image is too large")
		return
	}
	
	return
}

func ConvertImage(size string, filename string, data []byte) (odata []byte, err error) {
	var reader *bytes.Reader
	if len(data) > 10 * 1024 * 1024 {
		err = errors.New("Image is too large")
		return
	}
	reader = bytes.NewReader(data)
	return ConvertImageFile(size, filename, reader)
}

func AddImage(db *sqlx.DB, size string, filename string, data []byte) (id int64, err error) {
	var odata   []byte
	var res     sql.Result
	
	odata, err = ConvertImage(size, filename, data)
	if err != nil {
		return
	}
	
	res, err = db.Exec(`INSERT INTO images (FORMAT, DATA) VALUES (?, ?);`, convertFormat, odata)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	
	return
}

func GetImage(db *sqlx.DB, id int64) (data []byte, format string, err error) {
	err = db.QueryRowx(`SELECT FORMAT, DATA FROM images WHERE ID = ?;`, id).Scan(&format, &data)
	return
}

func AddImageFile(db *sqlx.DB, size string, filename string) (id int64, err error) {
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		return
	}
	return AddImage(db, size, filename, data)
}

func AddRoute(db *sqlx.DB, r *gin.Engine) {
	r.GET("/imgstore/:id", func(c *gin.Context) {
		var idS string
		var data []byte
		var format string
		var err error
		var found bool
		var id int
		idS, found = c.Params.Get("id")
		if !found || len(idS) == 0 {
			c.String(400, "Missing image id")
			return
		}
		id, err = strconv.Atoi(idS)
		if err != nil {
			c.String(400, "Invalid image id")
			return
		}
		data, format, err = GetImage(db, int64(id))
		if err != nil {
			c.String(404, "Image not found")
			return
		}
		c.Data(200, "image/" + format, data)
	})
}

func GetRoute(id int64) string {
	return "/imgstore/" + strconv.FormatInt(id, 10)
}
