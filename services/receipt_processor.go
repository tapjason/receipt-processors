package services

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"receipt-processor/models"

	"github.com/google/uuid"
)

type ReceiptProcessor struct {
	receipts map[string]models.Receipt
	mutex    sync.RWMutex
}

func NewReceiptProcessor() *ReceiptProcessor {
	return &ReceiptProcessor{
		receipts: make(map[string]models.Receipt),
	}
}

func (rp *ReceiptProcessor) ProcessReceipt(receipt models.Receipt) string {
	id := uuid.New().String()

	rp.mutex.Lock()
	rp.receipts[id] = receipt
	rp.mutex.Unlock()

	return id
}

func (rp *ReceiptProcessor) GetReceipt(id string) (models.Receipt, bool) {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	receipt, exists := rp.receipts[id]
	return receipt, exists
}

func (rp *ReceiptProcessor) CalculatePoints(receipt models.Receipt) int64 {
    var points int64 = 0

    // Rule 1: One point for every alphanumeric character in the retailer name
    alphanumericRegex := regexp.MustCompile("[a-zA-Z0-9]")
    points += int64(len(alphanumericRegex.FindAllString(receipt.Retailer, -1)))

    // Rule 2: 50 points if the total is a round dollar amount with no cents
    total, _ := strconv.ParseFloat(receipt.Total, 64)
    if math.Mod(total, 1.0) == 0 {
        points += 50
    }

    // Rule 3: 25 points if the total is a multiple of 0.25
    if math.Mod(total*100, 25) == 0 {
        points += 25
    }

    // Rule 4: 5 points for every two items on the receipt
    pairs := len(receipt.Items) / 2
    points += int64(pairs * 5)

    // Rule 5: If description length is multiple of 3, multiply price by 0.2 and round up
    for _, item := range receipt.Items {
        trimmedDesc := strings.TrimSpace(item.ShortDescription)
        if len(trimmedDesc) > 0 && len(trimmedDesc)%3 == 0 {
            price, _ := strconv.ParseFloat(item.Price, 64)
            points += int64(math.Ceil(price * 0.2))
        }
    }
    
    // Rule 6: 6 points if the day in the purchase date is odd
    purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
    if purchaseDate.Day()%2 == 1 {
        points += 6
    }

    // Rule 7: 10 points if purchase time is after 2:00pm and before 4:00pm
    purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
    hour := purchaseTime.Hour()
    minute := purchaseTime.Minute()
    purchaseTimeMinutes := hour*60 + minute
    
    // After 2:00pm means >= 14:00 (>=840 minutes)
    // Before 4:00pm means < 16:00 (<960 minutes)
    if purchaseTimeMinutes >= 14*60 && purchaseTimeMinutes < 16*60 {
        points += 10
    }

    return points
}