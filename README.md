# Audio Steganography LSB

Aplikasi full-stack modern untuk steganografi audio menggunakan metode Least Significant Bit (LSB) dan Parity. Aplikasi ini memungkinkan pengguna untuk menyembunyikan dan mengekstrak pesan rahasia dalam file audio secara aman sambil mempertahankan kualitas audio dan menyediakan berbagai algoritma embedding untuk kebutuhan keamanan dan kapasitas yang berbeda.

## 🎯 Deskripsi Program

Audio Steganography LSB adalah tool steganografi canggih yang memungkinkan pengguna untuk:

- **Menyembunyikan pesan rahasia** ke dalam file audio menggunakan teknik steganografi tingkat lanjut
- **Mengekstrak pesan tersembunyi** dari file audio yang telah dimodifikasi secara steganografis
- **Mendukung multiple metode steganografi**: metode LSB (Least Significant Bit) dan Parity
- **Mengenkripsi data yang di-embed** menggunakan algoritma kriptografi modern untuk keamanan yang lebih baik
- **Menghitung kapasitas embedding** untuk menentukan seberapa banyak data yang dapat disembunyikan dalam file audio
- **Memproses berbagai format audio** termasuk file WAV dan MP3
- **Menyediakan interface web modern** dengan desain bertema cyberpunk untuk interaksi pengguna yang intuitif

Aplikasi ini mengimplementasikan dua teknik steganografi utama:

1. **Metode LSB**: Memodifikasi bit paling tidak signifikan dari sample audio (1-4 bit dapat dikonfigurasi)
   - Kapasitas embedding tinggi
   - Processing cepat
   - Cocok untuk file rahasia berukuran besar

2. **Metode Parity**: Menyesuaikan paritas (jumlah bit genap/ganjil) dari sample audio
   - Ketahanan tinggi terhadap kompresi dan noise
   - Kapasitas lebih rendah tetapi keamanan lebih baik
   - Artifacts audio yang minimal

## 🛠 Tech Stack

### Backend
- **Go 1.25.1** - Bahasa backend utama
- **Gin Web Framework** - HTTP web framework untuk pengembangan API
- **Swagger/OpenAPI** - Interface dokumentasi dan testing API
- **MP3 Processing** - Dukungan format file audio via `hajimehoshi/go-mp3`
- **CORS Support** - Cross-origin resource sharing untuk integrasi frontend

### Frontend
- **React 19.1.1** - Framework frontend modern
- **TypeScript 5.8.3** - Pengembangan JavaScript yang type-safe
- **Vite 7.1.7** - Build tool dan development server yang cepat
- **Tailwind CSS 4.1.13** - CSS framework utility-first
- **Radix UI** - Library komponen UI yang accessible
- **React Hook Form** - Manajemen form dengan validasi

### Development Tools
- **ESLint** - Code linting dan formatting
- **Hot Reloading** - Development server dengan reload otomatis

## 📦 Dependencies

### Backend Dependencies

#### Core Framework
```
github.com/gin-gonic/gin v1.11.0           # Web framework
github.com/gin-contrib/cors v1.7.6         # CORS middleware
github.com/joho/godotenv v1.5.1            # Environment variables
```

#### Audio Processing
```
github.com/hajimehoshi/go-mp3 v0.3.4       # MP3 audio processing
```

#### API Documentation
```
github.com/swaggo/swag v1.16.6             # Swagger generation
github.com/swaggo/gin-swagger v1.6.1       # Gin Swagger integration
github.com/swaggo/files v1.0.1             # Swagger file serving
```

### Frontend Dependencies

#### Core Framework
```
react: ^19.1.1                             # React framework
react-dom: ^19.1.1                         # React DOM renderer
typescript: ~5.8.3                         # TypeScript support
vite: ^7.1.7                               # Build tool
```

#### UI Framework
```
tailwindcss: ^4.1.13                       # CSS framework
@radix-ui/react-*: 1.x.x                   # UI component library
lucide-react: ^0.454.0                     # Icon library
```

#### Form Management
```
react-hook-form: ^7.60.0                   # Form handling
@hookform/resolvers: ^3.10.0               # Form validation
zod: 3.25.67                               # Schema validation
```

#### Development Tools
```
@vitejs/plugin-react: ^5.0.3               # React Vite plugin
eslint: ^9.36.0                            # Code linting
typescript-eslint: ^8.44.0                 # TypeScript ESLint
```

## 🚀 Cara Menjalankan Program

### Prerequisites

Pastikan Anda telah menginstal software berikut di sistem Anda:

- **Go 1.25.1 atau lebih baru** - [Download Go](https://golang.org/dl/)
- **Node.js 18+ dan Bun** - [Download Node.js](https://nodejs.org/) | [Install Bun](https://bun.sh/)
- **Git** - [Download Git](https://git-scm.com/)
- **PowerShell** (Windows) atau shell yang kompatibel

### Instalasi & Setup

1. **Clone repository:**
   ```bash
   git clone https://github.com/Nerggg/Audio-Steganography-LSB.git
   cd Audio-Steganography-LSB
   ```

2. **Setup Backend:**
   ```powershell
   # Navigasi ke direktori backend
   cd backend
   
   # Install dependencies Go
   go mod download
   go mod tidy
   
   # Build aplikasi
   go build -o bin/audio-steganography-api.exe .
   ```

3. **Setup Frontend:**
   ```powershell
   # Navigasi ke direktori frontend
   cd ../frontend
   
   # Install dependencies menggunakan Bun
   bun install
   ```

### Menjalankan Aplikasi


**Backend:**
```powershell
cd backend
go run .
# Server akan berjalan di http://localhost:8080
```

**Frontend:**
```powershell
cd frontend
bun run dev
# Development server akan berjalan di http://localhost:5173
```

### Mengakses Aplikasi

- **Frontend Interface**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Dokumentasi API**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

### Development Commands

#### Frontend Commands
```bash
bun dev             # Start development server
bun build           # Build untuk production
bun preview         # Preview production build
bun lint            # Jalankan ESLint
```

### Production Build

**Backend:**
```powershell
cd backend
go build -ldflags="-w -s" -o audio-steganography-api.exe .
```

**Frontend:**
```bash
cd frontend
bun run build
```

### Troubleshooting

- **Konflik port**: Pastikan port 8080 (backend) dan 5173 (frontend) tersedia
- **Masalah CORS**: Periksa bahwa URL frontend diizinkan dalam konfigurasi CORS backend
- **Go modules**: Jalankan `go mod tidy` jika terjadi masalah dependency
- **Node modules**: Hapus `node_modules` dan `bun.lock`, lalu jalankan `bun install`

## 📁 Struktur Project

```
Audio-Steganography-LSB/
├── backend/                 # Go backend API
│   ├── handlers/           # HTTP request handlers
│   ├── models/            # Data models dan structures
│   ├── service/           # Business logic services
│   ├── docs/              # Dokumentasi API
│   └── main.go            # Application entry point
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── types/         # TypeScript type definitions
│   │   └── utils/         # Utility functions
│   └── public/            # Static assets
└── README.md              # Dokumentasi project
```

## 🔧 API Endpoints

- `GET /api/v1/health` - Health check
- `POST /api/v1/capacity` - Hitung kapasitas embedding
- `POST /api/v1/embed` - Embed pesan rahasia ke audio
- `POST /api/v1/extract` - Ekstrak pesan rahasia dari audio
- `GET /swagger/index.html` - Dokumentasi API

---

**Catatan**: Aplikasi ini dikembangkan sebagai bagian dari mata kuliah Kriptografi (IF4020) di Institut Teknologi Bandung.
