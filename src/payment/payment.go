package main

import (
	"fmt"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/wesleywillians/go-rabbitmq/queue"
	"io/ioutil"
	"net/url"
	"log"
	"net/http"
	uuid "github.com/satori/go.uuid"
)

type Result struct {
	Status string
}

type Order struct {
	ID uuid.UUID
	Coupon string
	CcNumber string
}

func NewOrder() Order {
	return Order{ID: uuid.NewV4()}
}

const (
	InvalidCoupon = "invalid"
	ValidCoupon = "valid"
	ConnectionError = "connection error"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
}

func main() {
	 messageChannel := make(chan amqp.Delivery)

	 rabbitMQ := queue.NewRabbitMQ()
	 ch := rabbitMQ.Connect()
	 defer ch.Close()

	 rabbitMQ.Consume(messageChannel)

	 for msg := range messageChannel {
		process(msg)
	 }
}


func process(msg amqp.Delivery) {
	order := NewOrder()
	json.Unmarshal(msg.Body, &order)

	resultCoupon := makeHttpCall("http://localhost:9092", order.Coupon)
	switch resultCoupon.Status {
	case InvalidCoupon:
		log.Println("Order: ", order.ID, ": ", resultCoupon.Status)
	case ConnectionError:
		msg.Reject(false)
		log.Println("Order: ", order.ID, ": ", resultCoupon.Status)
	case ValidCoupon:
		log.Println("Order: ", order.ID, ": ", resultCoupon.Status)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	coupon := r.PostFormValue("coupon")
	ccNumber := r.PostFormValue("ccNumber")

	resultCoupon := makeHttpCall("http://localhost:9092", coupon)

	result := Result{Status: "declined"}

	if ccNumber == "1" {
		result.Status = "approved"
	}

	if resultCoupon.Status == "invalid" {
		result.Status = "Coupon invalid"
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Fatal("Error processing json")
	}

	fmt.Fprintf(w, string(jsonData))

}


func makeHttpCall(microserviceUrl string, coupon string) Result {
	values := url.Values{}
	values.Add("coupon", coupon)

	res, err := http.PostForm(microserviceUrl, values)
	if err != nil {
		result := Result{Status: ConnectionError}
		return result
		//log.Fatal("Microservice payment out")
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// result := Result{Status: "Server is out"}
		// return result
		log.Fatal("Error processing result")
	}

	result := Result{}

	json.Unmarshal(data, &result)

	return result
}