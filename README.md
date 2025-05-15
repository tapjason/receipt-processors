
# Receipt Processor
  
A web service that processes receipts and calculates points based on specific rules.
  
## Prerequisites
  
- Go 1.18+
  
## Running the Application
  
1. Clone this repository
2. Navigate to the project directory
3. Run the application:
  
```go run main.go```

The server will start on port 8080.

## Testing
### Running Tests

To run the tests:

```go test -v ./tests```

### Test Coverage 
The test suite includes: 
1. **Unit Tests**: 
- Individual point calculation rules tested in isolation 
- Validation function tests
- Receipt processor service tests 
2. **Integration Tests**: 
- End-to-end API flow testing with HTTP requests 
- Validation of the example receipts from the requirements 
- Error handling scenarios (400/404 responses) 
3. **Rule-Specific Tests**: 
- Retailer name alphanumeric character points 
- Round dollar amount points 
- Multiple of 0.25 points
- Points for item pairs 
- Description length points 
- Odd day points 
- Afternoon time points 
4. **Edge Cases**: 
- Invalid receipt JSON 
- Missing required fields 
- Non-existent receipt IDs 

Each test is designed to verify specific aspects of the application in isolation, to ensure comprehensive coverage of all requirements.

## API Endpoints

### Process Receipt
- POST /receipts/process
- Request Body: Receipt JSON
- Response: JSON with receipt ID

### Get Points Total
- GET /receipts/{id}/points
- Response: JSON with points awarded

For example receipts, see the examples directory.

## Project Structure
receipt-processor/ 
├── main.go  # entry point
├── handlers/ 
│ └── receipt_handler.go  # API endpoint handlers
├── models/ 
│ └── receipt.go  # data models
├── services/ 
│ └── receipt_processor.go  # business logic
├── utils/ 
│ └── validation.go  # helper functions
├── tests/ 
│ └── receipt_processor_test.go  # test files
├── examples/ 
│ ├── simple-receipt.json 
│ └── morning-receipt.json 
├── go.mod 
├── go.sum 
└── README.md