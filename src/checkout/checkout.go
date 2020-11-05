package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/wesleywillians/go-rabbitmq/queue"
	"html/template"
	"log"
	"net/http"
)

type Result struct {
	Status string
}

type Order struct {
	Coupon string
	CcNumber string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/process", process)
	http.ListenAndServe(":9090", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/home.html"))
	t.Execute(w, Result{})
}

func process(w http.ResponseWriter, r *http.Request) {
	// log.Println(r.FormValue("coupon"))
	// log.Println(r.FormValue("cc-number"))
	//result := makeHttpCall("http://localhost:9091", r.FormValue("coupon"), r.FormValue("cc-number"))


	coupon := r.FormValue("coupon")
	ccNumber := r.FormValue("cc-number")

	order := Order{
		Coupon: coupon,
		CcNumber: ccNumber,
	}

	jsonOrder, err := json.Marshal(order)
	if err != nil {
		log.Fatal("Error parsing order json")
	}

	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()

	err = rabbitMQ.Notify(string(jsonOrder), "application/json", "orders_ex", "")
	if err != nil {
		log.Fatal("Error while trying to send rabbitmq message")
	}

	t := template.Must(template.ParseFiles("templates/process.html"))
	t.Execute(w, "")
}

//func makeHttpCall(microserviceUrl string, coupon string, ccNumber string) Result {
	//values := url.Values{}
	//values.Add("coupon", coupon)
	//values.Add("ccNumber", ccNumber)

	//retryClient := retryablehttp.NewClient()
	//retryClient.RetryMax = 5

	//res, err := retryClient.PostForm(microserviceUrl, values)
	//if err != nil {
		//result := Result{Status: "Server is out"}
		//return result
		//log.Fatal("Microservice payment out")
		//}

	//defer res.Body.Close()

	//data, err := ioutil.ReadAll(res.Body)
	//if err != nil {
		// result := Result{Status: "Server is out"}
		// return result
		//log.Fatal("Error processing result")
		//}

	//result := Result{}

	//json.Unmarshal(data, &result)

	//	return result
//}