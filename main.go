package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Mengambil kredensial dari Environment Variables yang disuntikkan GitHub Actions
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := "wastewatch_db" // Nama database biarkan diketik langsung
	
	// Format koneksi ke RDS MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPass, dbHost, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Peringatan: Gagal koneksi ke RDS - %v", err)
	} else {
		defer db.Close()
		fmt.Println("Berhasil inisialisasi koneksi RDS!")
	}

	// Route untuk halaman utama
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Selamat Datang di WasteWatch - Sistem Pelaporan Sampah")
	})

	// Route khusus untuk fitur Wajib S3 (Upload Foto Laporan)
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Gunakan method POST untuk upload", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("foto_sampah")
		if err != nil {
			http.Error(w, "Gagal membaca file upload", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Inisialisasi Sesi AWS S3 (Region Sydney)
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("ap-southeast-2"), 
		})
		
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
			Key:    aws.String("laporan/" + header.Filename),
			Body:   file,
		})
		
		if err != nil {
			http.Error(w, "Gagal upload ke S3: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Sukses! File %s berhasil disimpan ke S3.", header.Filename)
	})

	fmt.Println("Server WasteWatch berjalan di port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
