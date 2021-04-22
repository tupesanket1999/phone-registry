package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var phoneBook map[string]string

type Student struct {
	Name string
	Ph   string
}

var count uint64

func getCount(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "get count")
	fmt.Fprintln(w, count)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "add strudent")
	enableCors(&w)
	reqBody, _ := ioutil.ReadAll(r.Body)
	var student Student
	json.Unmarshal(reqBody, &student)

	if val, flag := phoneBook[student.Name]; flag {
		fmt.Fprintf(w, "Already there"+val)
	} else {
		phoneBook[student.Name] = student.Ph
		count++
		fmt.Fprintf(w, "Student added successfully")
	}
}

func getNameById(w http.ResponseWriter, r *http.Request) {
	//	json.NewEncoder(w).Encode(phoneBook)
	enableCors(&w)
	vars := mux.Vars(r)
	name := vars["name"]
	if val, flag := phoneBook[name]; flag {
		fmt.Fprintf(w, phoneBook[name])
	} else {
		fmt.Fprintf(w, "No Such student found"+val)
	}
}

func getAllNames(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(phoneBook)
}

func initDatabase() {
	loadFromDatabase()
	count = uint64(len(phoneBook))
}

var Mar = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
var Unmar = func(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func save(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Mar(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

func load(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		saveToDatabase()
		f, err := os.Open(path)
		fmt.Println(err)
		return Unmar(f, v)
	}
	defer f.Close()
	return Unmar(f, v)
}

func saveToDatabase() {
	for true {
		time.Sleep(60 * time.Second)
		fmt.Println("saved to disk")
		if err := save("./file.temp", phoneBook); err != nil {
			log.Fatalln(err)
		}
	}
}
func loadFromDatabase() {
	phoneBook = make(map[string]string)
	if err := load("./file.temp", &phoneBook); err != nil {
		log.Fatalln(err)
	}

}

func initServer() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.Handle("/", http.FileServer(http.Dir("./static")))

	myRouter.HandleFunc("/search/{name}", getNameById)
	myRouter.HandleFunc("/count", getCount)
	myRouter.HandleFunc("/all", getAllNames)
	myRouter.HandleFunc("/add", addStudent).Methods("POST", "OPTIONS")

	log.Fatal(http.ListenAndServe(":9876", myRouter))
}

func main() {
	initDatabase()
	go saveToDatabase()
	initServer()
}
