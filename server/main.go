package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"

	"github.com/tavo-wasd-gh/gocors"
)

type Cuenta struct {
	ID            string  `json:"id"`
	Nombre        string  `json:"nombre"`
	PGeneral      float32   `json:"p-general"`
	P1Servicios   float32   `json:"p1-servicios"`
	P1Suministros float32   `json:"p1-suministros"`
	P1Bienes      float32   `json:"p1-bienes"`
	P1Validez     time.Time `json:"p1-validez"`
	P2Servicios   float32   `json:"p2-servicios"`
	P2Suministros float32   `json:"p2-suministros"`
	P2Bienes      float32   `json:"p2-bienes"`
	P2Validez     time.Time `json:"p2-validez"`
	TEEU          bool      `json:"teeu"`
	COES          bool      `json:"coes"`
}

type Servicios struct {
	ID uint16 `json:"id"`
	Emitido time.Time `json:"emitido"`
	Cuenta string `json:"cuenta"`
	Detalle string `json:"detalle"`
	MontoBruto float32 `json:"monto-bruto"`
	MontoIVA float32 `json:"monto-iva"`
	MontoDesc float32 `json:"monto-desc"`
	JustifServ string `json:"justif-serv"`
	ProvNom string `json:"prov-nom"`
	ProvCed string `json:"prov-ced"`
	ProvDir string `json:"prov-dir"`
	ProvEmail string `json:"prov-email"`
	ProvTel string `json:"prov-tel"`
	ProvBanco string `json:"prov-banco"`
	ProvIBAN string `json:"prov-iban"`
	JustifProv string `json:"justif-prov"`
	COES bool `json:"coes"`
	GecoSol bool `json:"geco-sol"`
	GecoOCS bool `json:"geco-ocs"`
	PorEjecutar time.Time `json:"por-ejecutar"`
	Ejecutado time.Time `json:"ejecutado"`
	Pagado time.Time `json:"pagado"`
	Notas string `json:"notas"`
}

type Suministros struct {
	ID uint16 `json:"id"`
	Emitido time.Time `json:"emitido"`
	Cuenta string `json:"cuenta"`
	Desglose json.RawMessage `json:"desglose"`
	MontoBruto float32 `json:"monto-bruto"`
	JustifSum string `json:"justif-sum"`
	COES bool `json:"coes"`
	Geco string `json:"geco"`
	Notas string `json:"notas"`
}

type Bienes struct {
	ID uint16 `json:"id"`
	Emitido time.Time `json:"emitido"`
	Cuenta string `json:"cuenta"`
	Detalle string `json:"detalle"`
	MontoBruto float32 `json:"monto-bruto"`
	MontoIVA float32 `json:"monto-iva"`
	MontoDesc float32 `json:"monto-desc"`
	JustifBien string `json:"justif-bien"`
	ProvNom string `json:"prov-nom"`
	ProvCed string `json:"prov-ced"`
	ProvDir string `json:"prov-dir"`
	ProvEmail string `json:"prov-email"`
	ProvTel string `json:"prov-tel"`
	ProvBanco string `json:"prov-banco"`
	ProvIBAN string `json:"prov-iban"`
	JustifProv string `json:"justif-prov"`
	COES bool `json:"coes"`
	GecoSol bool `json:"geco-sol"`
	GecoOC bool `json:"geco-oc"`
	Recibido time.Time `json:"recibido"`
	Notas string `json:"notas"`
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatalf("Fatal: Missing env variables")
        }

	http.HandleFunc("/api/dashboard", handleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
        signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

        go func() {
		log.Println("Log: Running on :" + port + "...")
                if err := http.ListenAndServe(":" + port, nil); err != nil {
			log.Fatalf("Fatal: Failed to start on port %s: %v", port, err)
                }
        }()

        <-stop

	log.Println("Log: Shutting down...")
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	time.Sleep(200 * time.Millisecond)

	id := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/dashboard/"), "/", 2)[0]

	htmlTemplate, err := os.ReadFile("views/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to read template file", http.StatusInternalServerError)
		return
	}

	cuenta := Cuenta{
		ID: id,
		Nombre: "Asociación de Estudiantes de Ingeniería Mecánica",
		PGeneral: 0.00,
		P1Servicios: 1000000.00,
		P1Suministros: 1000000.00,
		P1Bienes: 1000000.00,
		P1Validez: time.Now().Add(10),
		P2Servicios: 1000000.00,
		P2Suministros: 1000000.00,
		P2Bienes: 1000000.00,
		P2Validez: time.Now().Add(10),
		TEEU: false,
		COES: true,
	}

	data := struct {
		Cuenta Cuenta
		// Solicitud Solicitud
	}{
		Cuenta: cuenta,
		// Solicitud: solicitud,
	}

	filledTemplate, err := fill(string(htmlTemplate), data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(filledTemplate)
}

func fill(htmlTemplate string, data interface{}) ([]byte, error) {
        template, err := template.New("").Parse(htmlTemplate)
        if err != nil {
                return nil, fmt.Errorf("failed to parse template: %w", err)
        }

        var filled bytes.Buffer
        if err := template.Execute(&filled, data); err != nil {
                return nil, fmt.Errorf("failed to execute template: %w", err)
        }

        return filled.Bytes(), nil
}
