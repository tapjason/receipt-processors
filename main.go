package main

import (
	"log"
	"net/http"

	"receipt-processor/handlers"
	"receipt-processor/services"

	"github.com/gorilla/mux"
)

func main() {
	receiptProcessor := services.NewReceiptProcessor()
	receiptHandler := handlers.NewReceiptHandler(receiptProcessor)

	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", receiptHandler.ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", receiptHandler.GetPoints).Methods("GET")

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}