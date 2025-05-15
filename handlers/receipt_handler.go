package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"receipt-processor/models"
	"receipt-processor/services"

	"github.com/gorilla/mux"
)

type ReceiptHandler struct {
	processor *services.ReceiptProcessor
}

func NewReceiptHandler(processor *services.ReceiptProcessor) *ReceiptHandler {
	return &ReceiptHandler{
		processor: processor,
	}
}

func (h *ReceiptHandler) ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt models.Receipt

	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	// Validate receipt fields
	if !isValidReceipt(receipt) {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	id := h.processor.ProcessReceipt(receipt)

	response := models.ReceiptResponse{ID: id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ReceiptHandler) GetPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	receipt, exists := h.processor.GetReceipt(id)
	if !exists {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	points := h.processor.CalculatePoints(receipt)

	response := models.PointsResponse{Points: points}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to validate receipt according to the requirements
func isValidReceipt(receipt models.Receipt) bool {
	// Basic validation
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) == 0 {
		return false
	}

	// Validate retailer
	retailerRegex := regexp.MustCompile(`^[\w\s\-&]+$`)
	if !retailerRegex.MatchString(receipt.Retailer) {
		return false
	}

	// Validate purchase date
	_, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		return false
	}

	// Validate purchase time
	_, err = time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		return false
	}

	// Validate total
	totalRegex := regexp.MustCompile(`^\d+\.\d{2}$`)
	if !totalRegex.MatchString(receipt.Total) {
		return false
	}

	// Validate each item
	for _, item := range receipt.Items {
		if item.ShortDescription == "" || !totalRegex.MatchString(item.Price) {
			return false
		}
	}

	return true
}