package main

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Define the available pack sizes
var packSizes = []int{250, 500, 1000, 2000, 5000}

// Calculate the minimum number of packs required to fulfill the order
func calculatePacks(itemCount int) map[int]int {
	result := make(map[int]int)
	calculatePacksRec(itemCount, result, len(packSizes)-1)

	return result
}

func calculatePacksRec(itemCount int, packs map[int]int, packIndex int) {
	packageSize := packSizes[packIndex]
	packageCount := itemCount / packageSize

	if packageCount > 0 {
		packs[packageSize] = packageCount
		itemCount %= packageSize
		if itemCount == 0 {
			return
		}
	}

	if itemCount < packSizes[0] {
		packs[packSizes[0]] = 1
		return
	}

	rightDistance := math.Abs(float64(itemCount) - float64(packageSize))
	packIndex -= 1
	leftDistance := math.Abs(float64(itemCount) - float64(packSizes[packIndex]))

	if packageCount == 0 && rightDistance < leftDistance {
		packs[packageSize] = 1
	} else if packageCount == 0 && itemCount > packSizes[packIndex] {
		packs[packageSize] = 1
	} else {
		calculatePacksRec(itemCount, packs, packIndex)
	}
}

func calculatePackHandler(c *gin.Context) {
	itemCountStr := c.Param("itemCount")
	itemCount, _ := strconv.Atoi(itemCountStr)

	packs := calculatePacks(itemCount)

	c.JSON(http.StatusOK, gin.H{"itemCount": itemCount, "packs": packs})
}

func main() {
	router := gin.Default()
	router.GET("/order/:itemCount", calculatePackHandler)
	router.Run(":8080")
}
