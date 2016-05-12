package main

import (
	"github.com/drone/routes"
	"log"
    "encoding/json"
	"net/http"
 	"strings"
	"strconv"
)

//To Hold Json generated by calling getUUID()
type UUID struct{
	Uuids []string `json:"uuids"`
}
//To store the details of the Student to be inserted in AddStudent()
type Student struct{
	StudentId int `json:"studentid"`
	RegClasses []int `json:"regclasses"`
	StudentName string `json:"studentname"`
}
//To store details of student that will be sent in response to GetStudentName()
type RowReturn struct{
	Id string `json:"id"`
	Key int `json:"key"`
	Value string `json:"value"`
}
//To store details of student that will be sent in response to GetStudentEnrolled()
type RowReturnEnrolled struct{
	Id string `json:"id"`
	Key int `json:"key"`
	Value []int `json:"value"`
}
//To Hold the json response when calling CouchDB's studentname api
type StudentNameAPI struct{
	Total_rows int `json:"total_rows"`
	Offset int `json:"offset"`
	Rows []struct{
			Id string `json:"id"`
			Key int `json:"key"`
			Value string `json:"value"`
		}`json:"rows"`
}
//To Hold the json response when calling CouchDB's studentenrolled api
type StudentEnrolledAPI struct{
	Total_rows int `json:"total_rows"`
	Offset int `json:"offset"`
	Rows []struct{
			Id string `json:"id"`
			Key int `json:"key"`
			Value []int `json:"value"`
		}`json:"rows"`
}


var uniqueid UUID
var studentNameAPI StudentNameAPI
var studentEnrolledAPI StudentEnrolledAPI
var BaseUrl string
var rowreturn RowReturn
var rowReturnEnrolled RowReturnEnrolled

//Get Unique UUID from CouchDB
func getUUID() string{
	response,_ := http.Get(BaseUrl+"/_uuids")
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&uniqueid)
	if err != nil {
		panic(err)
	}
	return uniqueid.Uuids[0]
}

//Helper Funtion does PUT for new student insert
func doPut(body string,uuid string){
    Url:= BaseUrl+"/studentlist/"+uuid
	request, _ := http.NewRequest("PUT", Url, strings.NewReader(body))
	client := &http.Client{}
	client.Do(request)
}

//Helper Function does GET for Student Names
func doGetName(id string){
	var Url string
	if id=="" {
		Url=BaseUrl+"/studentlist/_design/getlistdata/_view/studentname"
	}
	if id!="" {
	Url=BaseUrl+"/studentlist/_design/getlistdata/_view/studentname?key=+"+string(id)
	}
	response,_ := http.Get(Url)
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&studentNameAPI)
	if err != nil {
		panic(err)
	}
}

// Helper Function does GET for Student Enrolled
func doGetEnroll(id string){
	var Url string
	if id=="" {
		Url=BaseUrl+"/studentlist/_design/getlistdata/_view/studentenrolled"
	}
	if id!="" {
	Url=BaseUrl+"/studentlist/_design/getlistdata/_view/studentenrolled?key=+"+string(id)
	}
	response,_ := http.Get(Url)
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&studentEnrolledAPI)
	if err != nil {
		panic(err)
	}
}

//Main AddStudent function - REST
//Accepts a StudentId, RegisteredClasses[], Student Name and Insters into CouchDB
func AddStudent(rw http.ResponseWriter, r *http.Request) {
	var student Student
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&student)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"Json not correct"}`))
	}
	if err == nil {
		a, _ := json.Marshal(student)
		doPut(string([]byte(a)),getUUID())
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(""))
	}
}

//Main GetStudentName function - REST
//Accepts a StudentID and Returns Student Name
func GetStudentName(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get(":id")
	doGetName(id)
	rowreturn.Value= studentNameAPI.Rows[0].Value
	rowreturn.Key= studentNameAPI.Rows[0].Key
	rowreturn.Id= studentNameAPI.Rows[0].Id
	a, _ := json.Marshal(rowreturn)
	log.Println("Done")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(a))
}

//Main GetAllStudent function - REST
//Returns all StudentIDs and their Names
func GetAllStudent(rw http.ResponseWriter, r *http.Request) {
	doGetName("")
	a, _ := json.Marshal(studentNameAPI)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(a))
}

//Main GetStudentEnrolled function - REST
//Accepts a StudentID and returns all the classes they are enrolled in
func GetStudentEnrolled(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get(":id")
	doGetEnroll(id)
	rowReturnEnrolled.Value= studentEnrolledAPI.Rows[0].Value
	rowReturnEnrolled.Key= studentEnrolledAPI.Rows[0].Key
	rowReturnEnrolled.Id= studentEnrolledAPI.Rows[0].Id
	a, _ := json.Marshal(rowReturnEnrolled)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(a))
}

//Main CheckStudentValid function - REST
//Accepts StudentID and returns if student exists in the Database
func CheckStudentValid(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get(":id")
	doGetName(id)
	var a string
	if len(studentNameAPI.Rows)>0 {
		a=`{"status":"yes"}`
	}

	if len(studentNameAPI.Rows)==0 {
		a=`{"status":"no"}`
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(a))
}


func main(){

	//BASE URL for CouchDB. Curling this URL should give a welcome message
	BaseUrl=""

	//REST Config begins
			mux := routes.New()
			mux.Get("/studentname/:id", GetStudentName)
			mux.Get("/checkstudentvalid/:id", CheckStudentValid)
			mux.Get("/studentenrolled/:id", GetStudentEnrolled)
			mux.Get("/allstudent", GetAllStudent)
			mux.Post("/addstudent",AddStudent)
			http.Handle("/", mux)
			log.Println("REST has been set up: "+strconv.Itoa(3000))
			log.Println("Listening...")
			http.ListenAndServe(":"+strconv.Itoa(3000), nil)
	//REST Config end
}