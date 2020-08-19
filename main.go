package main

import (
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

/*
 * Model
 */
type User struct {
	ID       int
	Email    string
	Username string `gorm:"unique_index"`
}

/*
 * Initialize DB
 */
func db_init() {
	connectionStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("USER_NAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open("postgres", connectionStr)

	if err != nil {
		panic("filaed to conenct database")
	}

	defer db.Close()

	db.LogMode(true)

	// migration
	db.AutoMigrate(&User{})
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	//config.AllowOrigins = []string{"http://localhost:3001"}
	config.AllowHeaders = []string{"Origin", " Authorization"}
	r.Use(cors.New(config))

	db_init()

	api := r.Group("/api")

	//settion
	api.POST("/auth", authHandler)
	api.GET("/profile", getProfile)

	r.Run(":3001")
}

//POST /api/auth
func authHandler(c *gin.Context) {
	db, err := gorm.Open("postgres",
		"user=postgres password=postgres dbname=postgres sslmode=disable")

	if err != nil {
		panic("filaed to conenct database")
	}

	defer db.Close()

	var user User
	username := string(c.Request.FormValue("username"))

	//user exist
	ok, err := ExistsUserByName(username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	if ok {
		c.JSON(http.StatusOK, "ok")
		return
	}

	user.Username = string(c.Request.FormValue("username"))
	//user.Picture = string(c.Request.FormValue("picture"))
	user.UID = string(c.Request.FormValue("uid"))

	c.Bind(&user)
	db.Create(&user)
	c.JSON(200, "ok")
}

// GET: /profile
func getProfile(c *gin.Context) {

	db, err := gorm.Open("postgres",
		"user=postgres password=postgres dbname=postgres sslmode=disable")

	if err != nil {
		panic("filaed to conenct database")
	}

	defer db.Close()

	// firebase
	opt := option.WithCredentialsFile("your json key")
	app, err := firebase.NewApp(context.Background(), nil, opt)

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	authHeader := c.Request.Header.Get("Authorization")
	idToken := strings.Replace(authHeader, "Bearer ", "", 1)

	//tokenの検証
	token, err := client.VerifyIDToken(context.Background(), idToken)

	var user User
	if err := db.Where("uid= ?", token.UID).Find(&user).Error; err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
	} else {
		c.JSON(200, user)
	}
}
