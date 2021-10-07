package main

import (
	"myModule/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
)

func main() {
	r := gin.Default()

	r.GET("/blocks", func(c *gin.Context) {
		strBlocks := c.DefaultQuery(model.LIMIT, "20")
		varify := validate.Map(map[string]interface{}{model.LIMIT: strBlocks})
		varify.StringRule(model.LIMIT, "required|isNumber|min:1|max:20")

		if !varify.Validate() {
			c.JSON(http.StatusBadRequest, gin.H{"error": varify.Errors.One()})
			return
		}

		nBlocks, _ := strconv.Atoi(varify.GetSafe(model.LIMIT).(string))

		result, err := getLatestBlocks(nBlocks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getLatestBlocks(numLatestBlocks int) (model.ResponseBlocks, error) {
	return model.ResponseBlocks{
		Blocks: []model.BlockBase{
			{
				Num:        0,
				Hash:       "hash1",
				Time:       123456789,
				ParentHash: "",
			},
			{
				Num:        2,
				Hash:       "hash2",
				Time:       123456790,
				ParentHash: "hash1",
			},
		},
	}, nil
}
