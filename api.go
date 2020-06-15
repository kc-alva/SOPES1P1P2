package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type principal struct {
	Ejecucion     int         `json:"Ejecucion"`
	Suspendidos   int         `json:"Suspendidos"`
	Detenidos     int         `json:"Detendios"`
	Zombie        int         `json:"Zombie"`
	Total         int         `json:"Total"`
	ListaProcesos allProcesos `json:"ListaProc"`
}

type proceso struct {
	ID     string      `json:"ID"`
	PPID   string      `json:"PPID"`
	Nombre string      `json:"Nombre"`
	Estado string      `json:"Estado"`
	RAM    float64     `json:"RAM"`
	Hijos  allProcesos `json:"Hijos"`
}

type allProcesos []proceso
type allEvents []event

var events = allEvents{
	{
		ID:          "1",
		Title:       "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	},
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}

func getProcesos(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	carpetas, err := ioutil.ReadDir("/proc")

	if err != nil {
		log.Fatal(err)
	}

	totalejec := 0
	totalsuspe := 0
	totaldeten := 0
	totalzombie := 0

	var procesos = allProcesos{}
	var mensaje principal

	for _, carpeta := range carpetas {
		if carpeta.IsDir() {
			r, _ := regexp.Compile("[0-9]+")
			if !r.MatchString(carpeta.Name()) {
				continue
			}

			var nuevoProceso proceso
			nuevoProceso.ID = carpeta.Name()

			stat, err := ioutil.ReadFile("/proc/" + nuevoProceso.ID + "/stat")
			if err != nil {
				log.Fatal(err)
			}
			statm, err := ioutil.ReadFile(("/proc/") + nuevoProceso.ID + "/statm")
			if err != nil {
				log.Fatal(err)
			}

			contenidoStatm := strings.Split(string(statm), " ")
			contenido := strings.Split(string(stat), " ")
			ram, err := strconv.ParseFloat(contenidoStatm[1], 64)
			ram = ram * 4
			ram = (ram * 100) / 6003512
			nuevoProceso.Nombre = strings.Replace(contenido[1], "(", "", -1)
			nuevoProceso.Nombre = strings.Replace(nuevoProceso.Nombre, ")", "", -1)
			nuevoProceso.Estado = getEstado(contenido[2])
			nuevoProceso.RAM = ram
			nuevoProceso.PPID = contenido[3]
			procesos = append(procesos, nuevoProceso)

			switch contenido[2] {
			case "R":
				totalejec++
				break
			case "S":
				totalsuspe++
				break
			case "T":
				totaldeten++
				break
			case "Z":
				totalzombie++
				break
			}
		}
	}

	mensaje.Ejecucion = totalejec
	mensaje.Detenidos = totaldeten
	mensaje.Suspendidos = totalsuspe
	mensaje.Zombie = totalzombie
	mensaje.Total = totalejec + totaldeten + totalsuspe + totalzombie
	mensaje.ListaProcesos = procesos
	println(mensaje.Suspendidos)
	json.NewEncoder(w).Encode(mensaje)
}

func getEstado(s string) string {
	switch s {
	case "R":
		return "Running"
	case "S":
		return "Sleeping"
	case "D":
		return "Waiting"
	case "Z":
		return "Zombie"
	case "T":
		return "Stopped"
	case "t":
		return "Tracing stop"
	case "W":
		return "Paging"
	case "X":
	case "x":
		return "Dead"
	case "K":
		return "Wakekill"
	case "P":
		return "Parked"
	}

	return ""
}

func killProc(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	procID := mux.Vars(r)["id"]
	out, err := exec.Command("kill", procID).Output()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Proceso eliminado")
	output := string(out[:])
	fmt.Println(w, output)
}

func currentPercent(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	//get item value from json body
	idle0, total0 := getCPUSample()
	time.Sleep(500 * time.Millisecond)
	idle1, total1 := getCPUSample()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Dato1 float64 `json:"porcentaje"`
	}{Dato1: cpuUsage})
}

func getCPUSample() (idle, total uint64) {

	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			return
		}
	}
	return
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func ramData(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	memoTot, memoCons := getRAMSample()
	totalMB := (memoTot / 1024)
	consumedMB := (memoCons / 1024)
	memPercent := (((float64(memoCons) / 1024) * 100) / (float64(memoTot) / 1024))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Dato1 int     `json:"total"`
		Dato2 int     `json:"consumida"`
		Dato3 float64 `json:"porcentaje"`
	}{Dato1: totalMB, Dato2: consumedMB, Dato3: memPercent})
}

func getRAMSample() (memoriaTotal, memoriaConsumida int) {

	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	cnt := 0
	var memTotal string
	var memFree string
	a := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	for _, line := range lines {

		cnt++
		if cnt == 1 {

			b := a.FindAllString(line, 1)
			memTotal = b[0]
			fmt.Printf("tests - %s\n", b[0])
		} else if cnt == 2 {
			c := a.FindAllString(line, 1)
			memFree = c[0]
		}

	}
	memoriaTotal, err = strconv.Atoi(memTotal)
	memoriaLibre, err := strconv.Atoi(memFree)
	memoriaConsumida = memoriaTotal - memoriaLibre
	return
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/proc", getProcesos).Methods("GET")
	router.HandleFunc("/kill/{id}", killProc).Methods("GET")
	router.HandleFunc("/cpu", currentPercent).Methods("GET")
	router.HandleFunc("/ram", ramData).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
