package main

import (
   "github.com/gin-gonic/gin"
   swaggerFiles "github.com/swaggo/files"
   ginSwagger "github.com/swaggo/gin-swagger"
   docs "github.com/Nerggg/Audio-Steganography-LSB/backend/docs"
   "github.com/Nerggg/Audio-Steganography-LSB/backend/service"
)

// @BasePath /api/v1

func main()  {
   r := gin.Default()
   docs.SwaggerInfo.BasePath = "/api/v1"
   v1 := r.Group("/api/v1")
   {
      v1.POST("/api/capacity",service.CalculateCapacityHandler)
      v1.POST("/api/embed",service.EmbedHandler)
   }
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   r.Run(":8085")

}
