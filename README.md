# GO-IMGSTORE

Facility to save images in SQL.

## Go documentation

    package imgstore // import "github.com/harkaitz/go-imgstore"
    
    imgstore uses imagemagick to resize and convert the images supplied by the
    user and stores them in a database (tested in PostgreSQL).
    
    func AddImage(db *sqlx.DB, size string, filename string, data []byte) (id int64, err error)
    func AddImageFile(db *sqlx.DB, size string, filename string, file io.Reader) (id int64, err error)
    func AddImageFilename(db *sqlx.DB, size string, filename string) (id int64, err error)
    func AddRoute(db *sqlx.DB, r *gin.Engine)
    func ConvertImage(size string, filename string, data []byte) (odata []byte, err error)
    func ConvertImageFile(size string, filename string, file io.Reader) (odata []byte, err error)
    func CreateTables(db *sqlx.DB) (err error)
    func GetImage(db *sqlx.DB, id int64) (data []byte, format string, err error)
    func GetRoute(id int64) string
    type Image struct{ ... }

