package main

import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    docs "github.com/Nerggg/Audio-Steganography-LSB/backend/docs"
   "github.com/Nerggg/Audio-Steganography-LSB/backend/controller"
)

// @BasePath /api/v1

func main()  {
   r := gin.Default()
   docs.SwaggerInfo.BasePath = "/api/v1"
   v1 := r.Group("/api/v1")
   {
      eg := v1.Group("/example")
      {
         eg.GET("/helloworld",controller.Helloworld)
         eg.GET("/helloworld2",controller.Helloworld2)
      }
   }
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   r.Run(":8080")

}
