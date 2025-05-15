package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	
	"receipt-processor/models"
	"receipt-processor/services"
	"receipt-processor/handlers"
)

func TestSimpleReceiptCalculation(t *testing.T) {
	processor := services.NewReceiptProcessor()
	
	// Create test data from simple-receipt.json
	receipt := models.Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-02",
		PurchaseTime: "13:13",
		Total:        "1.25",
		Items: []models.Item{
			{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
		},
	}
	
	points := processor.CalculatePoints(receipt)
	
	// Expected points breakdown:
	// - 6 points for "Target" (6 alphanumeric characters)
	// - 25 points for total being a multiple of 0.25
	// - 0 points for day not being odd
	// Total: 31 points
	expectedPoints := int64(31)
	
	if points != expectedPoints {
		t.Errorf("Simple receipt calculation failed: got %d points, expected %d", points, expectedPoints)
	}
}

func TestMorningReceiptCalculation(t *testing.T) {
	processor := services.NewReceiptProcessor()
	
	// Create test data from morning-receipt.json
	receipt := models.Receipt{
		Retailer:     "Walgreens",
		PurchaseDate: "2022-01-02",
		PurchaseTime: "08:13",
		Total:        "2.65",
		Items: []models.Item{
			{ShortDescription: "Pepsi - 12-oz", Price: "1.25"},
			{ShortDescription: "Dasani", Price: "1.40"},
		},
	}
	
	points := processor.CalculatePoints(receipt)
    
    // Expected points breakdown:
    // - 9 points for "Walgreens" (9 alphanumeric characters)
    // - 0 points for total not being a round dollar amount
    // - 0 points for total not being a multiple of 0.25
    // - 5 points for 2 items (1 pair)
    // - 1 point for "Dasani" (6 characters, multiple of 3, 1.40 * 0.2 = 0.28, rounded up to 1)
    // - 0 points for purchase time not between 2-4 PM
    // - 0 points for purchase date not being odd
    // Total: 15 points
    expectedPoints := int64(15)
    
    if points != expectedPoints {
        t.Errorf("Morning receipt calculation failed: got %d points, expected %d", points, expectedPoints)
    }
}

func TestFullAPIFlow(t *testing.T) {
	processor := services.NewReceiptProcessor()
	handler := handlers.NewReceiptHandler(processor)
	
	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", handler.ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", handler.GetPoints).Methods("GET")

	// Test both receipts through the API
	testCases := []struct {
		name           string
		receiptJSON    string
		expectedPoints int64
	}{
		{
			name: "Simple Receipt",
			receiptJSON: `{
				"retailer": "Target",
				"purchaseDate": "2022-01-02", 
				"purchaseTime": "13:13",
				"total": "1.25",
				"items": [
					{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}
				]
			}`,
			expectedPoints: 31,
		},
		{
			name: "Morning Receipt",
			receiptJSON: `{
				"retailer": "Walgreens",
				"purchaseDate": "2022-01-02",
				"purchaseTime": "08:13",
				"total": "2.65",
				"items": [
					{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
					{"shortDescription": "Dasani", "price": "1.40"}
				]
			}`,
			expectedPoints: 15,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Process receipt
			req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBufferString(tc.receiptJSON))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("Processing receipt failed: got status %d, expected 200", rr.Code)
			}

			var processResponse models.ReceiptResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &processResponse); err != nil {
				t.Fatalf("Failed to parse process response: %v", err)
			}

			pointsURL := "/receipts/" + processResponse.ID + "/points"
			req, _ = http.NewRequest("GET", pointsURL, nil)
			rr = httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("Getting points failed: got status %d, expected 200", rr.Code)
			}

			var pointsResponse models.PointsResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &pointsResponse); err != nil {
				t.Fatalf("Failed to parse points response: %v", err)
			}
			if pointsResponse.Points != tc.expectedPoints {
				t.Errorf("Points calculation failed: got %d points, expected %d", 
                           pointsResponse.Points, tc.expectedPoints)
			}
		})
	}
}

func TestReadmeExamples(t *testing.T) {
	processor := services.NewReceiptProcessor()
	
	// Example 1 from README
	receipt1 := models.Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []models.Item{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
			{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
			{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
			{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
			{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
		},
		Total: "35.35",
	}
	
	points1 := processor.CalculatePoints(receipt1)
	expectedPoints1 := int64(28)
	
	if points1 != expectedPoints1 {
		t.Errorf("README example 1 failed: got %d points, expected %d", points1, expectedPoints1)
	}
	
	// Example 2 from README
	receipt2 := models.Receipt{
		Retailer:     "M&M Corner Market",
		PurchaseDate: "2022-03-20",
		PurchaseTime: "14:33",
		Items: []models.Item{
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
		},
		Total: "9.00",
	}
	
	points2 := processor.CalculatePoints(receipt2)
	expectedPoints2 := int64(109)
	
	if points2 != expectedPoints2 {
		t.Errorf("README example 2 failed: got %d points, expected %d", points2, expectedPoints2)
	}
}

func TestErrorHandling(t *testing.T) {
	processor := services.NewReceiptProcessor()
	handler := handlers.NewReceiptHandler(processor)
	
	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", handler.ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", handler.GetPoints).Methods("GET")
	
	// Test invalid receipt JSON
	t.Run("Invalid Receipt JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBufferString(`{invalid json}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Invalid JSON should return 400 Bad Request, got %d", rr.Code)
		}
	})
	
	// Test missing required fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBufferString(`{"retailer": "Test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Missing required fields should return 400 Bad Request, got %d", rr.Code)
		}
	})
	
	// Test non-existent receipt ID
	t.Run("Non-existent Receipt ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/receipts/non-existent-id/points", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusNotFound {
			t.Errorf("Non-existent receipt ID should return 404 Not Found, got %d", rr.Code)
		}
	})
}

// Individual Point Rules Tests
func TestIndividualPointRules(t *testing.T) {
	processor := services.NewReceiptProcessor()
	
	// Test Rule 1: One point for every alphanumeric character in the retailer name
	t.Run("Rule 1: Retailer Name Points", func(t *testing.T) {
		baseReceipt := models.Receipt{
			Retailer:     "X",
			PurchaseDate: "2022-01-02",
			PurchaseTime: "12:00",
			Items:        []models.Item{{ShortDescription: "Item", Price: "1.10"}},
			Total:        "1.10",
		}
		basePoints := processor.CalculatePoints(baseReceipt)
		
		testReceipt := baseReceipt
		testReceipt.Retailer = "Tar-get & Store"
		testPoints := processor.CalculatePoints(testReceipt)
		
		expectedDiff := int64(10)
		actualDiff := testPoints - basePoints
		if actualDiff != expectedDiff {
			t.Errorf("Retailer name points failed: got point difference of %d, expected %d", 
				actualDiff, expectedDiff)
		}
	})
	
	// Test Rule 2: 50 points if the total is a round dollar amount with no cents
	t.Run("Rule 2: Round Dollar Amount", func(t *testing.T) {
		baseReceipt := models.Receipt{
			Retailer:     "X",
			PurchaseDate: "2022-01-02",
			PurchaseTime: "12:00",
			Items:        []models.Item{{ShortDescription: "Item", Price: "5.01"}},
			Total:        "5.01",
		}
		basePoints := processor.CalculatePoints(baseReceipt)
		
		testReceipt := baseReceipt
		testReceipt.Total = "5.00"
		testReceipt.Items[0].Price = "5.00"
		testPoints := processor.CalculatePoints(testReceipt)
		
		expectedDiff := int64(75)
		actualDiff := testPoints - basePoints
		if actualDiff != expectedDiff {
			t.Errorf("Round dollar points incorrect: difference was %d, expected %d", 
				actualDiff, expectedDiff)
		}
		
		quarterReceipt := baseReceipt
		quarterReceipt.Total = "5.25"
		quarterReceipt.Items[0].Price = "5.25"
		quarterPoints := processor.CalculatePoints(quarterReceipt)
		
		expectedRoundVsQuarterDiff := int64(50)
		actualRoundVsQuarterDiff := testPoints - quarterPoints
		if actualRoundVsQuarterDiff != expectedRoundVsQuarterDiff {
			t.Errorf("Round dollar vs quarter dollar points incorrect: difference was %d, expected %d", 
				actualRoundVsQuarterDiff, expectedRoundVsQuarterDiff)
		}
	})
	
	// Test Rule 3: 25 points if the total is a multiple of 0.25
	t.Run("Rule 3: Multiple of 0.25", func(t *testing.T) {
		baseReceipt := models.Receipt{
			Retailer:     "X",
			PurchaseDate: "2022-01-02",
			PurchaseTime: "12:00",
			Items:        []models.Item{{ShortDescription: "Item", Price: "5.23"}},
			Total:        "5.23",
		}
		basePoints := processor.CalculatePoints(baseReceipt)
		
		testReceipt := baseReceipt
		testReceipt.Total = "5.25"
		testPoints := processor.CalculatePoints(testReceipt)
		
		expectedDiff := int64(25)
		actualDiff := testPoints - basePoints
		if actualDiff != expectedDiff {
			t.Errorf("Multiple of 0.25 points incorrect: difference was %d, expected %d", 
				actualDiff, expectedDiff)
		}
	})
	
	// Test Rule 4: 5 points for every two items on the receipt
	t.Run("Rule 4: Points for Item Pairs", func(t *testing.T) {
		baseItem := models.Item{ShortDescription: "Item", Price: "1.01"}
		
		for i := 1; i <= 5; i++ {
			items := make([]models.Item, i)
			for j := 0; j < i; j++ {
				items[j] = baseItem
			}
			
			expectedItemPairPoints := int64((i / 2) * 5)
			
			totalPrice := 1.01 * float64(i)
			totalStr := strconv.FormatFloat(totalPrice, 'f', 2, 64)
			
			receipt := models.Receipt{
				Retailer:     "X",
				PurchaseDate: "2022-01-02",
				PurchaseTime: "12:00",
				Items:        items,
				Total:        totalStr,
			}
			
			points := processor.CalculatePoints(receipt)
			constPoints := int64(1)

			expectedTotalPoints := constPoints + expectedItemPairPoints
			if points != expectedTotalPoints {
				t.Errorf("Item pair points for %d items incorrect: got %d points, expected %d", 
					i, points, expectedTotalPoints)
			}
		}
	})
	
	// Test Rule 5: Description length multiple of 3
	t.Run("Rule 5: Description Length Points", func(t *testing.T) {
		baseReceipt := models.Receipt{
			Retailer:     "X",
			PurchaseDate: "2022-01-02",
			PurchaseTime: "12:00",
			Items: []models.Item{
				{ShortDescription: "ABCD", Price: "5.00"},
			},
			Total: "5.00",
		}
		basePoints := processor.CalculatePoints(baseReceipt)
		
		testCases := []struct {
			description  string
			price        string
			expectedDiff int64
		}{
			{"ABC", "5.00", 1},
			{"ABCDEF", "5.00", 1},
			{"ABC", "4.90", 1},
			{"ABC", "0.50", 1},
			{"ABC", "10.00", 2},
			{"   ABC   ", "5.00", 1},
		}
		
		for _, tc := range testCases {
			testReceipt := baseReceipt
			testReceipt.Items[0].ShortDescription = tc.description
			testReceipt.Items[0].Price = tc.price
			
			testPoints := processor.CalculatePoints(testReceipt)
			actualDiff := testPoints - basePoints
			if actualDiff != tc.expectedDiff {
				t.Errorf("Description '%s' with price %s points incorrect: difference was %d, expected %d", 
					tc.description, tc.price, actualDiff, tc.expectedDiff)
			}
		}
	})
	
	// Test Rule 6: 6 points if the day in the purchase date is odd
	t.Run("Rule 6: Odd Day Points", func(t *testing.T) {
		testCases := []struct {
			date     string
			expected int64
		}{
			{"2022-01-01", 6},
			{"2022-01-02", 0},
			{"2022-01-31", 6},
			{"2022-01-30", 0},
		}
		
		for _, tc := range testCases {
			receipt := models.Receipt{
				Retailer:     "X",
				PurchaseDate: tc.date,
				PurchaseTime: "12:00",
				Items:        []models.Item{{ShortDescription: "Item", Price: "1.01"}},
				Total:        "1.01",
			}
			
			points := processor.CalculatePoints(receipt)
			basePoints := int64(1)
			
			expectedPoints := basePoints + tc.expected
			if points != expectedPoints {
				t.Errorf("Odd day points incorrect for date %s: got %d points, expected %d", 
					tc.date, points, expectedPoints)
			}
		}
	})
	
	// Test Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	t.Run("Rule 7: Afternoon Time Points", func(t *testing.T) {
		testCases := []struct {
			time     string
			expected int64
		}{
			{"13:59", 0},
			{"14:00", 10},
			{"14:30", 10},
			{"15:59", 10},
			{"16:00", 0},
			{"17:00", 0},
		}
		
		for _, tc := range testCases {
			receipt := models.Receipt{
				Retailer:     "X",
				PurchaseDate: "2022-01-02",
				PurchaseTime: tc.time,
				Items:        []models.Item{{ShortDescription: "Item", Price: "1.01"}},
				Total:        "1.01",
			}
			
			points := processor.CalculatePoints(receipt)
			basePoints := int64(1)
			
			expectedPoints := basePoints + tc.expected
			if points != expectedPoints {
				t.Errorf("Time points incorrect for %s: got %d points, expected %d", 
					tc.time, points, expectedPoints)
			}
		}
	})
}