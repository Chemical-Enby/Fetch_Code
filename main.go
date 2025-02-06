/*
	Author: Christian (Sapphire) Godard
	Date: 2/6/2025
	File: main.go
	Description: Setup receipt API for entering receipts and getting their points
*/

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Item holds the description and price of items that are associated with a receipt
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// Receipt holds the retailer, purchase date, purchase time, total, and items purchased associated with a receipt
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

var receipts = make(map[string]Receipt)

/*
Given a specific receipt it will score it based on a multitude of criteria relating to name of the retailer, total
amount on purchase, how many items were purchased, length of item descriptions, purchase date, and purchase time
*/
func receiptPoints(receipt Receipt) (points int) {
	points = 0

	// Alphanumeric retailer character check
	for _, char := range receipt.Retailer {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			points++
		}
	}

	// Receipt total checks
	receiptTotal, _ := strconv.ParseFloat(receipt.Total, 64)

	if math.Mod(receiptTotal, 1.00) == 0 {
		points += 50
	}

	if math.Mod(receiptTotal, 0.25) == 0 {
		points += 25
	}

	// Receipt items check
	points += 5 * (len(receipt.Items) / 2)

	// Item description check
	for _, currItem := range receipt.Items {
		if len(strings.Trim(currItem.ShortDescription, " "))%3 == 0 {
			currPrice, _ := strconv.ParseFloat(currItem.Price, 64)
			points += int(math.Ceil(currPrice * 0.2))
		}
	}

	// Date check
	receiptDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)

	if err != nil {
		log.Println("Couldn't parse receipt date of " + receipt.PurchaseDate)
		return 0
	}

	if receiptDate.Day()%2 != 0 {
		points += 6
	}

	// Time check
	receiptTime, err := time.Parse("15:04", receipt.PurchaseTime)

	if err != nil {
		log.Println("Couldn't parse receipt time of " + receipt.PurchaseTime)
		return 0
	}

	if receiptTime.Hour() >= 14 && receiptTime.Hour() < 16 {
		if receiptTime.Hour() == 14 && receiptTime.Minute() > 0 {
			points += 10
		} else {
			points += 10
		}
	}

	return points
}

/*
Checks if receipt exists, and it returns an error if not found. Otherwise, we get the points of the receipt and
return that.
*/
func getReceiptPoints(c *gin.Context) {
	receiptId := c.Param("id")

	if currReceipt, exists := receipts[receiptId]; exists {
		c.IndentedJSON(http.StatusOK, gin.H{"points": receiptPoints(currReceipt)})
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Receipt not found"})
}

/*
Creates receipt and adds it to our receipt collection if it is valid
*/
func postReceipt(c *gin.Context) {
	var newReceipt Receipt

	if err := c.BindJSON(&newReceipt); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	receiptGuid := uuid.New()

	for {
		if _, exists := receipts[receiptGuid.String()]; exists {
			receiptGuid = uuid.New()
		} else {
			break
		}
	}

	receipts[receiptGuid.String()] = newReceipt
	c.IndentedJSON(http.StatusCreated, gin.H{"id": receiptGuid.String()})
}

/*
Create receipt API and run it on localhost:8080. Check if error occurred when starting it
*/
func main() {
	router := gin.Default()

	api := router.Group("/receipts")
	{
		api.GET("/:id/points", getReceiptPoints)
		api.POST("/process", postReceipt)
	}

	err := router.Run("localhost:8080")

	if err != nil {
		log.Fatal("Something BAD HAPPENED" + err.Error())
		return
	}
}
