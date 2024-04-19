package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Result struct {
	Id      string
	Journey string
}

type UserProfile struct {
	Id       string
	Name     string
	Location string
	X        float64
	Y        float64
	Distance string
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/InputProccessor", InputProccessor).Methods("POST")
	router.HandleFunc("/", healthCheck).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	})

	handler := c.Handler(router)
	log.Println("Started")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "It's working")
}

func InputProccessor(w http.ResponseWriter, r *http.Request) {
	var userProfile []UserProfile
	var numDrone int

	err := json.NewDecoder(r.Body).Decode(&userProfile)

	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Somethings went wrong")
	}

	requestFile, fileErr := os.Create("0-0-requestInfo.txt")
	if fileErr != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Somethings went wrong")
	}

	_, fileErr = requestFile.WriteString("id	   #	     x	     y	     w \n")
	for index, user := range userProfile {
		if user.Id == "-1" {
			numDrone = int(user.X)
		} else {
			x := fmt.Sprintf("%f", user.X)
			y := fmt.Sprintf("%f", user.Y)
			_, fileErr = requestFile.WriteString(strconv.Itoa(index) + "   " + strconv.Itoa(index) + "	     " + x + "	     " + y + "	     0  \n")
		}
	}
	if fileErr != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Somethings went wrong")
	} else {
		w.WriteHeader(http.StatusOK)
	}
	defer requestFile.Close()

	Proccessing(numDrone)

	var result []Result

	result = make([]Result, numDrone+1)
	methodList := OutputProccess(numDrone)
	log.Println(methodList)
	for index, method := range methodList {
		var tmp Result
		tmp.Id = strconv.Itoa(index)
		tmp.Journey = method
		result[index] = tmp
	}

	json.NewEncoder(w).Encode(result)
}

func insertNumDrone(numDrone int) {
	input, err := os.ReadFile("ouput_get_tau.txt")
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "numDrone") {
			lines[i] = "numDrone " + strconv.Itoa(numDrone)
		}
	}
	output := strings.Join(lines, "\n")
	err = os.WriteFile("ouput_get_tau.txt", []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func Proccessing(numDrone int) {
	cmd := exec.Command("InputProccessor/get-tau.exe")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Lỗi khi chạy file exe:", err)
		fmt.Println("Output của lệnh:", string(output))

	} else {
		fmt.Println("Chạy file exe thành công")
		fmt.Println("Output của lệnh:", string(output))
	}
	insertNumDrone(numDrone)

	cmd = exec.Command("PathCalculator/Logistics_test.exe")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Lỗi khi chạy file exe:", err)
		fmt.Println("Output của lệnh:", string(output))

	} else {
		fmt.Println("Chạy file exe thành công")
		fmt.Println("Output của lệnh:", string(output))
	}

}

func OutputProccess(numDrone int) []string {
	var methodList []string
	var drone []string
	var truck []string

	input, err := os.ReadFile("sol_info.txt")
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for _, line := range lines {
		if strings.Contains(line, "D") {
			drone = strings.Split(line, " ")
			drone[0] = getLastString(drone[0])
		}
		if strings.Contains(line, "T") {
			truck = strings.Split(line, " ")
			truck[0] = getLastString(truck[0])
		}
	}
	methodList = make([]string, numDrone+1)
	for _, t := range truck {
		if t[0] >= '0' && t[0] <= '9' {
			methodList[0] += string(t)
			methodList[0] += " "
		}
		log.Println("t ", t)
	}

	i := 1
	for _, d := range drone {
		if d[0] >= '0' && d[0] <= '9' {
			methodList[i] += string(d)
			methodList[i] += " "
		}
		if i >= numDrone {
			i = 1
		} else {
			i++
		}
		log.Println("d ", d)
	}
	log.Println(truck)

	return methodList
}

func getLastString(str string) string {
	value := ""
	i := len(str) - 1

	for str[i] >= '0' && str[i] <= '9' && i >= 0 {
		value = value + string(str[i])
		log.Println(str[i], " ", value, " ", i)
		i--
	}
	return value
}
