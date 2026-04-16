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

// Desain UI Halaman Utama (HTML + CSS Bootstrap)
const htmlTemplate = `
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WasteWatch - Sistem Pelaporan Sampah</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body class="bg-light">
    <div class="container mt-5" style="max-width: 600px;">
        <div class="text-center mb-4">
            <h1 class="text-success fw-bold">♻️ WasteWatch</h1>
            <p class="text-muted">Sistem Pelaporan Penumpukan Sampah Kota</p>
        </div>
        
        <div class="card shadow-sm">
            <div class="card-body p-4">
                <form action="/upload" method="POST" enctype="multipart/form-data">
                    <div class="mb-3">
                        <label class="form-label fw-bold">Nama Pelapor</label>
                        <input type="text" name="nama" class="form-control" placeholder="Masukkan nama Anda" required>
                    </div>
                    <div class="mb-3">
                        <label class="form-label fw-bold">Lokasi Tumpukan Sampah</label>
                        <textarea name="lokasi" class="form-control" rows="2" placeholder="Contoh: Pinggir jalan raya..." required></textarea>
                    </div>
                    <div class="mb-4">
                        <label class="form-label fw-bold">Unggah Foto Bukti</label>
                        <input type="file" name="foto_sampah" class="form-control" accept="image/*" required>
                        <small class="text-muted">Foto ini akan otomatis disimpan ke dalam AWS S3 Bucket.</small>
                    </div>
                    <button type="submit" class="btn btn-success w-100 fw-bold">Kirim Laporan</button>
                </form>
            </div>
        </div>
    </div>
</body>
</html>
`

func main() {
	// 1. Inisialisasi Koneksi ke AWS RDS MySQL
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := "wastewatch_db"
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPass, dbHost, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Gagal koneksi ke RDS: %v", err)
	} else {
		defer db.Close()
		fmt.Println("Berhasil inisialisasi koneksi RDS MySQL!")
	}

	// 2. Route Halaman Beranda (Menampilkan Web UI)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlTemplate)
	})

	// 3. Route untuk Memproses Upload Data ke AWS S3
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Batas maksimal ukuran file (10 MB)
		r.ParseMultipartForm(10 << 20) 
		nama := r.FormValue("nama")
		lokasi := r.FormValue("lokasi")

		file, header, err := r.FormFile("foto_sampah")
		if err != nil {
			http.Error(w, "Gagal membaca foto", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Buka Sesi ke AWS S3
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

		// Balasan Halaman Sukses dengan Konfirmasi HTML
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
			<div class="container mt-5 text-center" style="max-width: 600px;">
				<div class="card shadow-sm p-5">
					<h2 class="text-success fw-bold">✅ Laporan Berhasil Dikirim!</h2>
					<p class="mt-3">Terima kasih <b>%s</b>. Laporan tumpukan sampah di <b>%s</b> telah kami terima.</p>
					<p class="text-muted small">File foto <i>%s</i> telah diamankan ke server penyimpanan Cloud AWS S3.</p>
					<a href="/" class="btn btn-outline-success mt-4">Kembali ke Beranda</a>
				</div>
			</div>
		`, nama, lokasi, header.Filename)
	})

	fmt.Println("Server WasteWatch berjalan di port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
