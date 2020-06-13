package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

/*
type item struct {
	UsedCPU   float32 `json:"CPU"`
	OtherData int     `json:"otherData"`
}

//post is a post that gets in the post method
type Post struct {
	CPU  float64 `json:"cpu"`
	Body string  `json:"body"`
}

var posts []Post = []Post{}
*/
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/cpu", currentPercent).Methods("GET")
	http.ListenAndServe(":5000", router)
}

func currentPercent(w http.ResponseWriter, r *http.Request) {

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
