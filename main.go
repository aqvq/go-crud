package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// User 结构体
// 注意点
// 1. 结构体里面的变量（Name）必须是首字母大写
// 2. gorm指定类型 json表示json接受的时候的名称 binding required表示必须传入
type User struct {
	gorm.Model
	Name    string `gorm:"type:varchar(20); not null" json:"name" binding:"required"`
	State   string `gorm:"type:varchar(20); not null" json:"state" binding:"required"`
	Phone   string `gorm:"type:varchar(20); not null" json:"phone" binding:"required"`
	Email   string `gorm:"type:varchar(40); not null" json:"email" binding:"required"`
	Address string `gorm:"type:varchar(200); not null" json:"address" binding:"required"`
}

func main() {

	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:shang@tcp(127.0.0.1:3306)/go-crud?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	err2 := db.AutoMigrate(&User{})
	if err2 != nil {
		return
	}
	// 接口
	r := gin.Default()

	// 业务码约定
	// 正确返回200
	// 错误返回400

	// 增
	r.POST("/user/create", func(c *gin.Context) {
		//var user User
		var user User
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(200, gin.H{
				"msg":  "添加失败",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			// 数据库操作
			db.Create(&user)
			c.JSON(200, gin.H{
				"msg":  "添加成功",
				"code": 200,
				"data": user,
			})
		}
	})
	// 删
	// 1. 找到对应的id
	// 2. 判断id是否存在
	// 3. 从数据库中删除
	// 3. 返回，id没有找到
	r.DELETE("/user/delete", func(c *gin.Context) {
		var user User
		// 接收id
		id := c.Query("id")
		//fmt.Println(id)
		// 判断id是否存在
		result := db.Where("id = ?", id).Take(&user)
		if result.Error != nil {
			c.JSON(200, gin.H{
				"msg":  "没有找到id，删除失败",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			// id存在的情况，则删除，不存在则报错
			// 操作数据库删除
			db.Delete(&user)
			c.JSON(200, gin.H{
				"msg":  "删除成功",
				"code": 200,
				"data": user,
			})
		}

	})
	// 改
	r.PUT("/user/update", func(c *gin.Context) {
		id := c.Query("id")
		var user User
		result := db.Where("id = ?", id).Take(&user)
		if result.Error != nil {
			c.JSON(200, gin.H{
				"msg":  "用户id没有找到",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			err := c.ShouldBindJSON(&user)
			if err != nil {
				c.JSON(200, gin.H{
					"msg":  "修改失败",
					"code": 400,
					"data": gin.H{},
				})
			} else {
				db.Where("id = ?", id).Updates(&user)
				c.JSON(200, gin.H{
					"msg":  "修改成功",
					"code": 200,
					"data": user,
				})
			}
		}
	})

	// 查 (条件查询，全部查询、分页查询)
	// 条件查询
	r.GET("/user/read", func(c *gin.Context) {
		// 获取路径参数
		name := c.Query("name")
		//fmt.Println(name)
		var users []User
		// 查询数据库
		db.Where("name = ?", name).Find(&users)
		// 判断是否查询到数据
		//fmt.Println(users)
		if len(users) == 0 {
			id := c.Query("id")
			db.Where("id = ?", id).Find(&users)
			if len(users) != 0 {
				c.JSON(200, gin.H{
					"msg":  "查询成功",
					"code": 200,
					"data": users,
				})
				return
			}
			c.JSON(200, gin.H{
				"msg":  "没有查询到数据",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功",
				"code": 200,
				"data": users,
			})
		}
	})
	// 全部查询
	r.GET("/user/list", func(c *gin.Context) {
		var users []User
		// 查询全部数据 查询分页数据
		pageSize := cast.ToInt(c.Query("pageSize"))
		pageNum := cast.ToInt(c.Query("pageNum"))
		if pageSize == 0 {
			pageSize = -1
		}
		if pageNum == 0 {
			pageNum = -1
		}
		offsetVal := (pageNum - 1) * pageSize
		if pageNum == -1 && pageSize == -1 {
			offsetVal = -1
		}
		// 查询数据库
		var total int64
		//fmt.Println(pageSize)
		//fmt.Println(offsetVal)
		db.Model(&User{}).Count(&total).Limit(pageSize).Offset(offsetVal).Find(&users)
		//fmt.Println(total)
		if len(users) == 0 {
			c.JSON(200, gin.H{
				"msg":  "没有查询到数据",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功",
				"code": 200,
				"data": gin.H{
					"list":     users,
					"total":    total,
					"pageNum":  pageNum,
					"pageSize": pageSize,
				},
			})
		}
	})
	// 端口号
	PORT := "8080"
	r.Run(":" + PORT)
}
