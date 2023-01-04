package main

import (
	"encoding/hex"
	"fmt"
	tokenapi "gin_demo/src"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

type User struct {
	gorm.Model
	Name      string `gorm:"type:varchar(20);not null"`
	Telephone string `gorm:"type:varchar(11);not null"`
	Password  string `gorm:"type:varchar(20);not null"`
}

func main() {
	db := InitDB()
	// 服务创建
	ginServer := gin.Default()

	// 加载静态页面
	ginServer.LoadHTMLGlob("templates/*")

	// 加载资源文件
	ginServer.Static("/static", "./static")

	// 加载视频
	//ginServer.LoadHTMLFiles("static/ys.mp4")

	ginServer.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hello": "the max hw",
		})
	})

	ginServer.GET("api/auth/register", func(context *gin.Context) {
		name := context.PostForm("name")
		telephone := context.PostForm("telephone")
		password := context.PostForm("password")

		if len(telephone) != 11 {
			context.JSON(http.StatusUnprocessableEntity, map[string]any{
				"code": 422,
				"msg":  "手机号必须为11位",
			})
			return
		}
		if len(password) < 6 {
			context.JSON(http.StatusUnprocessableEntity,
				gin.H{
					"code": 422,
					"msg":  "密码不应小于6位",
				})
			return
		}

		if len(name) == 0 {
			name = RandomString(10)
		}

		fmt.Println(name, telephone, password)

		if isTelephoneexist(db, telephone) {
			context.JSON(http.StatusUnprocessableEntity, gin.H{
				"code": 422,
				"msg":  "用户已经存在",
			})
			return
		}

		newUser := User{
			Name:      name,
			Telephone: telephone,
			Password:  password,
		}
		db.Create(&newUser)

		//	返回结果
		context.JSON(http.StatusOK, gin.H{
			"msg": "注册成功",
		})
	})
	videoSer := ginServer.Group("video")
	{
		videoSer.GET("/:videoname", func(context *gin.Context) {
			videoname := context.Param("videoname")
			filename := path.Join("./static", videoname)
			isExit, _ := PathExists(filename)
			if isExit {
				context.File(filename)
			} else {
				context.JSON(http.StatusNotFound, gin.H{})
			}
		})
		videoSer.GET("/", func(context *gin.Context) {
			context.HTML(http.StatusOK, "video.html", gin.H{})
		})
	}

	userSer := ginServer.Group("/user")
	{
		userSer.GET("/login", func(context *gin.Context) {
			context.HTML(http.StatusOK, "index.html", map[string]any{
				"msg": "后端返回数据",
			})
		})
		userSer.GET("/login/:username/:password", tokenapi.Login)
		userSer.GET("/verify/:token", tokenapi.Verify)
		userSer.GET("/refresh/:token", tokenapi.Refresh)
		userSer.GET("/sayHello/:token", tokenapi.SayHello)

		userSer.POST("/", func(context *gin.Context) {
			username := context.PostForm("username")
			//password := context.PostForm("password")
			context.HTML(http.StatusOK, "index.html", gin.H{
				"username": username,
			})
		})

	}

	ginServer.GET("/retest", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "https://www.baidu.com/")
	})
	ginServer.Run(":8888") // 监听并在 0.0.0.0:8080 上启动服务
}

// PathExists 判断一个文件或文件夹是否存在
// 输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	uLen := n
	b := make([]byte, uLen)
	rand.Read(b)
	rand_str := hex.EncodeToString(b)[0:uLen]
	return rand_str
}

func InitDB() *gorm.DB {
	//driverName := "mysql"
	//host := "locakhost"
	//port := "3306"
	//database := "ginessential"
	//username := "root"
	//password := "toor"
	//charset := "utf-8"
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True",
	//	username,
	//	password,
	//	host,
	//	port,
	//	database,
	//	charset,
	//)
	//db, err := gorm.Open(driverName, args)
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("userinfo.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema
	db.AutoMigrate(&User{})
	return db
}

func isTelephoneexist(db *gorm.DB, telephone string) bool {
	var user User
	db.Where("telephone = ?", telephone).First(&user)
	if user.ID != 0 {
		return true
	}
	return false
}
