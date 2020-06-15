package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/cpu", cpuData).Methods("GET")
	router.HandleFunc("/ram", ramData).Methods("GET")
	http.ListenAndServe(":5000", router)
}

func cpuData(w http.ResponseWriter, r *http.Request) {

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
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

func ramData(w http.ResponseWriter, r *http.Request) {

	//get item value from json body
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
