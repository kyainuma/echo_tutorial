package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type (
  User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
  }

  CustomValidator struct {
    validator *validator.Validate
  }
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// type User struct {
// 	ID string `param:"id" query:"id" form:"id" json:"id" xml:"id"`
// }

type UserDTO struct {
	Name string
	Email string
	IsAdmin bool
}

type CustomBinder struct {}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) (err error) {
	db := new(echo.DefaultBinder)
	if err := db.Bind(i, c); err != echo.ErrUnsupportedMediaType {
		return err
	}

	// カスタムバインドを実装
	return
}

type CustomContext struct {
	echo.Context
}

func (c *CustomContext) Foo() {
	println("foo")
}

func (c *CustomContext) Bar() {
	println("bar")
}

type CustomStruct struct {
	Timestamp Timestamp `query:"timestamp"`
}

type Timestamp time.Time

func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(time.RFC3339, src)
	*t = Timestamp(ts)
	return err
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(err)
	errorPage := fmt.Sprintf("%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
}

func main() {
	e := echo.New()
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c}
			return next(cc)
		}
	})
	e.Validator = &CustomValidator{validator: validator.New()}
	e.POST("/validate-users", func(c echo.Context) (err error) {
		u := new(User)
		if err = c.Bind(u); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err = c.Validate(u); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, u)
	})

	initRouting(e)
	// Root level middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${time_rfc3339} ${level}")
		l.SetLevel(log.INFO)
	}

	e.GET("/logger", func(c echo.Context) error {
		e.Logger.Info("logger func is called")
		return c.String(http.StatusOK, "logger!")
	})

	// Group level middleware
	g := e.Group("/admin")
	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "joe" && password == "secret" {
			return true, nil
		}
		return false, nil
	}))

	// Route level middleware
	track := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			println("request to /users")
			return next(c)
		}
	}
	e.GET("/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "/users")
	}, track)

	e.GET("/query-param", func(c echo.Context) error {
		name := c.QueryParam("name")
		return c.String(http.StatusOK, name)
	})

	e.GET("users/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.String(http.StatusOK, name)
	})

	// h2s := &http2.Server{
	// 	MaxConcurrentStreams: 250,
	// 	MaxReadFrameSize: 1048576,
	// 	IdleTimeout: 10 * time.Second,
	// }
	// s := http.Server {
	// 	Addr: ":8080",
	// 	Handler: h2c.NewHandler(e, h2s),
	// 	ReadTimeout: 30 * time.Second,
	// 	TLSConfig: &tls.Config{},
	// }
	// if err := s.ListenAndServeTLS("server.crt", "server.key"); err != http.ErrServerClosed {
	// 	log.Fatal(err)
	// }
	e.Logger.Fatal(e.Start(":1323"))
}

func initRouting(e *echo.Echo) {
	e.GET("/hello", hello)
	e.GET("/users/:id", getUser)
	e.GET("/show", show)
	e.POST("/save", save)
	e.POST("/users", userSave)
	e.Static("/static", "assets")
	e.File("/", "public/index.html")
	e.GET("/api/search", search)
	e.GET("/context", context)
	e.GET("/parallel_context", parallelContext)
	e.GET("/write_cookie", writeCookie)
	e.GET("/read_cookie", readCookie)
	e.GET("/read_all_cookie", readAllCookies)
	e.POST("/form", formValue)
	e.GET("/timestamp", timestamp)
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World")
}

func getUser(c echo.Context) error {
	// var user User
	// err := c.Bind(&user); if err != nil {
	// 	return c.String(http.StatusBadRequest, "Bad Request")
	// }
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func show(c echo.Context) error {
	team := c.QueryParam("team")
	member := c.QueryParam("member")
	return c.String(http.StatusOK, "team:" + team + ", member:" + member)
}

func save(c echo.Context) error {
	name := c.FormValue("name")
	avatar, err := c.FormFile("avatar")
	if err != nil {
		return err
	}

	// Source
	src, err := avatar.Open()
	if err != nil {
		return err
	}

	// Destination
	dst, err := os.Create(avatar.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, "<b>Thank you! " + name + "</b>")
}

func userSave(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// セキュリティ向上のため別のStructにロードする
	user := UserDTO {
		Name: u.Name,
		Email: u.Email,
		IsAdmin: false, // バインドされるべきではないフィールドの公開を回避する
	}

	// executeSomeBusinessLogic(user)

	return c.JSON(http.StatusOK, user)
}

func search(c echo.Context) error {
	var opts struct {
		IDs []int64
		Active bool
	}
	length := int64(50) // デフォルト値

	// バインドの例外処理
	err := echo.QueryParamsBinder(c).
		Int64("length", &length).
		Int64s("ids", &opts.IDs).
		Bool("active", &opts.Active).
		BindError()

	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	return c.JSON(http.StatusOK, opts)
}

func context(c echo.Context) error {
	cc := c.(*CustomContext)
	cc.Foo()
	cc.Bar()
	return cc.String(200, "OK")
}

func parallelContext(c echo.Context) error {
	ca := make(chan string, 1)
	r := c.Request()
	method := r.Method

	go func() {
		fmt.Printf("Method: %s\n", method)

		ca <- "Hay!"
	}()

	select {
	case result := <-ca:
		return c.String(http.StatusOK, "Result: " + result)
	case <-c.Request().Context().Done():
		return nil
	}
}

func writeCookie(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "username"
	cookie.Value = "job"
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "write a cookie")
}

func readCookie(c echo.Context) error {
	cookie, err := c.Cookie("username")
	if err != nil {
		return err
	}
	fmt.Println(cookie.Name)
	fmt.Println(cookie.Value)
	return c.String(http.StatusOK, "read a cookie")
}

func readAllCookies(c echo.Context) error {
	for _, cookie := range c.Cookies() {
		fmt.Println(cookie.Name)
		fmt.Println(cookie.Value)
	}
	return c.String(http.StatusOK, "read all the cookies")
}

func formValue(c echo.Context) error {
	name := c.FormValue("name")
	return c.String(http.StatusOK, name)
}

func timestamp(c echo.Context) error {
	cs := &CustomStruct{}
	if err := c.Bind(cs); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, cs)
}