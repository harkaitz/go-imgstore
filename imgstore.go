// imgstore uses imagemagick to resize and convert the images
// supplied by the user and stores them in a database (tested in
// PostgreSQL).
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
	"io"
)

type Image struct {
	ID       int64   `db:"id"        sql:"AUTO_INCREMENT" `
	Format   string  `db:"format"    sql:"NOT NULL"       `
	Data     []byte  `db:"data"      sql:"NOT NULL"       `
}

// CreateTables creates the "images" table needed by imgstore.
func CreateTables (db *sqlx.DB) (err error) {
	_, err = db.Exec(`
	-- SQL --
	CREATE TABLE IF NOT EXISTS images (
	    id        SERIAL          NOT NULL PRIMARY KEY,
	    format    VARCHAR(16)     NOT NULL,
	    data      BYTEA           NOT NULL
	);
	-- SQL --`)
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

// ConvertImageFile converts an image supplied by the user to "png"
// with a maximun size of 512kB.
func ConvertImageFile(size string, filename string, file io.Reader) (odata []byte, err error) {
	var cmd    *exec.Cmd
	var stdout  bytes.Buffer
	var stderr  bytes.Buffer
	var suffix  string
	
	suffix = filepath.Ext(filename)
	switch suffix {
	case ".jpg",".jpeg",".png", ".JPG", ".JPEG", ".PNG":
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

// ConvertImage converts an image supplied by the user to "png"
// with a maximun size of 512kB.
func ConvertImage(size string, filename string, data []byte) (odata []byte, err error) {
	var reader *bytes.Reader
	if len(data) > 10 * 1024 * 1024 {
		err = errors.New("Image is too large")
		return
	}
	reader = bytes.NewReader(data)
	return ConvertImageFile(size, filename, reader)
}

// AddImageFile converts an image with "ConvertImageFile" and stores
// it in the "images" table.
func AddImageFile(db *sqlx.DB, size string, filename string, file io.Reader) (id int64, err error) {
	var odata []byte
	odata, err = ConvertImageFile(size, filename, file)
	if err != nil {
		return
	}
	return addImageRaw(db, odata)
}

// AddImage converts an image with "ConvertImageFile" and stores
// it in the "images" table. Returns an ID.
func AddImage(db *sqlx.DB, size string, filename string, data []byte) (id int64, err error) {
	var odata []byte
	odata, err = ConvertImage(size, filename, data)
	if err != nil {
		return
	}
	return addImageRaw(db, odata)
}


func addImageRaw(db *sqlx.DB, odata []byte) (id int64, err error) {
	var cmd string = `
	-- SQL --
	INSERT INTO images (format, data) VALUES ($1, $2) RETURNING id;
	-- SQL --`
	err = db.QueryRowx(cmd, convertFormat, odata).Scan(&id)
	if err != nil {
		log.Print("ERROR SQL: " + cmd)
		return
	}
	log.Printf("Added new image: %v\n", id)
	return
}

// GetImage reads the stored image from the database.
func GetImage(db *sqlx.DB, id int64) (data []byte, format string, err error) {
	var cmd string = `
	-- SQL --
	SELECT format, data FROM images WHERE id = $1;
	-- SQL --`
	err = db.QueryRowx(cmd, id).Scan(&format, &data)
	if err != nil {
		log.Print("ERROR SQL: " + cmd)
		return
	}
	return
}

// AddImageFilename converts an image with "ConvertImageFile" and stores
// it in the "images" table.
func AddImageFilename(db *sqlx.DB, size string, filename string) (id int64, err error) {
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		return
	}
	return AddImage(db, size, filename, data)
}

// AddRoute binds the "/imgstore/ID" route to the image store.
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

// GetRoute returns the route of an image ("/imgstore/ID").
func GetRoute(id int64) string {
	return "/imgstore/" + strconv.FormatInt(id, 10)
}

