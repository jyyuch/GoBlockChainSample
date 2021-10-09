package main

import (
	"myModule/model"
	"myModule/proxy"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
)

func main() {
	r := gin.Default()

	r.GET("/blocks", getBlocks)
	r.GET("/blocks/:id", getBlockById)
	// r.GET("/transaction/:txHash", getTranxByHash)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getBlocks(c *gin.Context) {
	// verify params
	strBlocks := c.DefaultQuery(model.LIMIT, "20")
	varify := validate.Map(map[string]interface{}{model.LIMIT: strBlocks})
	varify.StringRule(model.LIMIT, "required|isNumber|min:1|max:20")

	if !varify.Validate() {
		c.JSON(http.StatusBadRequest, gin.H{"error": varify.Errors.One()})
		return
	}

	// get latest n blocks
	nBlocks, _ := strconv.Atoi(varify.GetSafe(model.LIMIT).(string))
	result, err := proxy.EthGetLatestBlocks(uint64(nBlocks))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// return 200 ok
	c.JSON(http.StatusOK, result)
}

func getBlockById(c *gin.Context) {
	// verify params
	varify := validate.Map(map[string]interface{}{model.ID: c.Param(model.ID)})
	varify.StringRule(model.ID, "required|isNumber|min:0")

	if !varify.Validate() {
		c.JSON(http.StatusBadRequest, gin.H{"error": varify.Errors.One()})
		return
	}

	// get latest n blocks
	blockNum, _ := strconv.Atoi(varify.GetSafe(model.ID).(string))
	result, err := proxy.EthFetchBlockByNumber(uint64(blockNum))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// return 200 ok
	c.JSON(http.StatusOK, result)
}

// func getTranxByHash(c *gin.Context) {
// 	// verify params
// 	varify := validate.Map(map[string]interface{}{model.TRANX_HASH: c.Param(model.TRANX_HASH)})
// 	varify.StringRule(model.TRANX_HASH, "required|isNumber|min:0")

// 	if !varify.Validate() {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": varify.Errors.One()})
// 		return
// 	}

// 	// get latest n blocks
// 	blockNum, _ := strconv.Atoi(varify.GetSafe(model.ID).(string))
// 	result, err := proxy.EthFetchBlockByNumber(uint64(blockNum))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// return 200 ok
// 	c.JSON(http.StatusOK, result)
// }
