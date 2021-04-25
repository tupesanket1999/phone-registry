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
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/semaphore"
)

var phoneBook map[string]string

type Student struct {
	Name string
	Ph   string
}

var Rlck int32
var Wlck int32

var c uint64

//var m = &sync.RWMutex{}

// LOCK START
func readLock() {
	for atomic.LoadInt32(&Rlck) > 0 {
	}
	atomic.AddInt32(&Rlck, 1)
}

func readUnlock() {
	atomic.AddInt32(&Rlck, -1)
}

func writeLock() {
	for atomic.LoadInt32(&Rlck) > 0 && atomic.LoadInt32(&Wlck) > 0 {
	}
	atomic.AddInt32(&Wlck, 1)
}

func writeUnlock() {
	atomic.AddInt32(&Wlck, -1)
}

// LOCK ENDED
func getCount(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "get count")
	//	m.RLock()
	readLock()
	fmt.Fprintln(w, c)
	readUnlock()
	//	m.RUnlock()
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

var sem = semaphore.NewWeighted(int64(10))

func addStudent(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "add strudent")
	enableCors(&w)
	reqBody, _ := ioutil.ReadAll(r.Body)
	var student Student
	json.Unmarshal(reqBody, &student)

	//ctx := context.Background()

	//what if user is not accessing map for same student name
	//how can we find out if is wants to modify for same name
	//sync.map

	//	sem.Acquire(ctx, 1)

	//atomic int
	//read write locks
	writeLock()
	if val, flag := phoneBook[student.Name]; flag {
		fmt.Fprintf(w, "exists "+val+"\n")
	} else {
		//multiple users ?
		//locks
		//		mutex.Lock()

		phoneBook[student.Name] = student.Ph
		c++
		fmt.Fprintf(w, "Student "+student.Name+" added successfully\n")
		//		defer m.Unlock()
		defer writeUnlock()
	}
}

func getNameById(w http.ResponseWriter, r *http.Request) {
	//	json.NewEncoder(w).Encode(phoneBook)
	enableCors(&w)
	vars := mux.Vars(r)
	name := vars["name"]
	readLock()
	if val, flag := phoneBook[name]; flag {
		fmt.Fprintf(w, phoneBook[name])
	} else {
		fmt.Fprintf(w, "No Such student found"+val)
	}
	defer readUnlock()
}

func getAllNames(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(phoneBook)
}

func initDatabase() {
	loadFromDatabase()
	c = uint64(len(phoneBook))
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
		sv()
		f, err := os.Open(path)
		fmt.Println(err)
		return Unmar(f, v)
	}
	defer f.Close()
	return Unmar(f, v)
}
func sv() {
	fmt.Println("saved to disk")
	if err := save("./file.json", phoneBook); err != nil {
		log.Fatalln(err)
	}
}

func saveToDatabase() {
	for true {
		time.Sleep(60 * time.Second)
		fmt.Println("saved to disk")
		if err := save("./file.json", phoneBook); err != nil {
			log.Fatalln(err)
		}
	}
}
func loadFromDatabase() {
	phoneBook = make(map[string]string)
	if err := load("./file.json", &phoneBook); err != nil {
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
