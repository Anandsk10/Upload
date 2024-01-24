#upload document
package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type PhotoUploadRequest struct {
	EventID string `json:"event_id"`
	Photos  string `json:"photos"`
}

func main() {
	var err error
	DB, err = sql.Open("mysql", "bdms_staff_admin:sfhakjfhyiqundfgs3765827635@tcp(buzzwomendatabase-new.cixgcssswxvx.ap-south-1.rds.amazonaws.com:3306)/bdms_staff?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", UploadEventPhotos)
	http.ListenAndServe(":8080", nil)
}

func UploadEventPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusNotFound, "message": "Method not found", "success": false})
		return
	}

	var request PhotoUploadRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Invalid Request Body", "success": false})
		return
	}

	if request.EventID == "" || request.Photos == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Invalid Input passed", "success": false})
		return
	}

	// Decode base64 image data
	imgData, err := base64.StdEncoding.DecodeString(request.Photos)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Failed to decode image data", "success": false})
		return
	}

	// Generate unique image name
	imageName := time.Now().Format("20060102150405") + request.EventID + ".jpg"

	// Create the images folder if it doesn't exist
	err = os.MkdirAll("images", os.ModePerm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Failed to create images folder", "success": false, "error": err})
		return
	}

	// Save the image file
	imagePath := filepath.Join("images", imageName)
	err = ioutil.WriteFile(imagePath, imgData, os.ModePerm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Failed to save image file", "success": false, "error": err})
		return
	}

	// Update the database with the image path
	updateQuery := "UPDATE tbl_poa SET photo1 = ? WHERE id = ?"
	_, err = DB.Exec(updateQuery, imagePath, request.EventID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusBadRequest, "message": "Failed to update photo path in the database", "success": false, "error": err})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusOK, "message": "Photos Uploaded Successfully", "success": true})
}
