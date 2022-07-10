package routes

import (
	"github.com/chris92vr/mongodb-go-jwt/controllers"
	"github.com/gin-gonic/gin"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/signup", controllers.SignUp())
	incomingRoutes.POST("/users/login", controllers.Login())
	incomingRoutes.GET("/users/me", controllers.GetUser())
	incomingRoutes.POST("/users/logout", controllers.Logout())

}
