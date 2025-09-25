package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    docs "github.com/Nerggg/Audio-Steganography-LSB/backend/docs"
)

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context)  {
   g.JSON(http.StatusOK,"helloworld")
}

func main()  {
   r := gin.Default()
   docs.SwaggerInfo.BasePath = "/api/v1"
   v1 := r.Group("/api/v1")
   {
      eg := v1.Group("/example")
      {
         eg.GET("/helloworld",Helloworld)
      }
   }
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   r.Run(":8080")

}
