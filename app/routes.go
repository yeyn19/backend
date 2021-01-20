package app

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/leoleoasd/EduOJBackend/app/controller"
	"github.com/leoleoasd/EduOJBackend/app/middleware"
	"github.com/leoleoasd/EduOJBackend/base/config"
	"github.com/leoleoasd/EduOJBackend/base/log"
	"github.com/leoleoasd/EduOJBackend/base/utils"
	"net/http"
	"net/http/pprof"
)

func Register(e *echo.Echo) {
	utils.InitOrigin()
	e.Use(middleware.Recover)
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: utils.Origins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	api := e.Group("/api", middleware.Authentication)

	auth := api.Group("/auth", middleware.Auth)
	auth.POST("/login", controller.Login).Name = "auth.login"
	auth.POST("/register", controller.Register).Name = "auth.register"
	auth.GET("/email_registered", controller.EmailRegistered).Name = "auth.emailRegistered"

	admin := api.Group("/admin", middleware.Logged)
	admin.POST("/user",
		controller.AdminCreateUser, middleware.HasPermission("manage_user")).Name = "admin.user.createUser"
	admin.PUT("/user/:id",
		controller.AdminUpdateUser, middleware.HasPermission("manage_user")).Name = "admin.user.updateUser"
	admin.DELETE("/user/:id",
		controller.AdminDeleteUser, middleware.HasPermission("manage_user")).Name = "admin.user.deleteUser"
	admin.GET("/user/:id",
		controller.AdminGetUser, middleware.HasPermission("read_user")).Name = "admin.user.getUser"
	admin.GET("/users",
		controller.AdminGetUsers, middleware.HasPermission("read_user")).Name = "admin.user.getUsers"

	api.GET("/user/me", controller.GetMe, middleware.Logged).Name = "user.getMe"
	api.PUT("/user/me", controller.UpdateMe, middleware.Logged).Name = "user.updateMe"
	api.GET("/user/:id", controller.GetUser).Name = "user.getUser"
	api.GET("/users", controller.GetUsers).Name = "user.getUsers"

	api.POST("/user/change_password", controller.ChangePassword, middleware.Logged).Name = "user.changePassword"

	api.GET("/image/:id", controller.GetImage).Name = "image.getImage"
	api.POST("/image", controller.CreateImage, middleware.Logged).Name = "image.createImage"

	admin.POST("/problem",
		controller.CreateProblem, middleware.HasPermission("create_problem")).Name = "problem.createProblem"
	admin.PUT("/problem/:id",
		controller.UpdateProblem, middleware.HasPermission("update_problem", "problem")).Name = "problem.updateProblem"
	admin.DELETE("/problem/:id",
		controller.DeleteProblem, middleware.HasPermission("delete_problem", "problem")).Name = "problem.deleteProblem"

	api.GET("/problem/:id", controller.GetProblem).Name = "problem.getProblem"
	api.GET("/problems", controller.GetProblems).Name = "problem.getProblems"

	api.GET("/problem/:id/attachment_file", controller.GetProblemAttachmentFile).Name = "problem.getProblemAttachmentFile"

	admin.POST("/problem/:id/test_case",
		controller.CreateTestCase,
		middleware.HasPermission("update_problem", "problem")).Name = "problem.createTestCase"
	admin.PUT("/problem/:id/test_case/:test_case_id",
		controller.UpdateTestCase,
		middleware.HasPermission("update_problem", "problem")).Name = "problem.updateTestCase"
	admin.DELETE("/problem/:id/test_case/all",
		controller.DeleteTestCases,
		middleware.HasPermission("update_problem", "problem")).Name = "problem.deleteTestCases"
	admin.DELETE("/problem/:id/test_case/:test_case_id",
		controller.DeleteTestCase,
		middleware.HasPermission("update_problem", "problem")).Name = "problem.deleteTestCase"

	admin.GET("/problem/:id/test_case/:test_case_id/input_file",
		controller.GetTestCaseInputFile,
		middleware.HasPermission("read_problem_secret", "problem")).Name = "problem.getTestCaseInputFile"
	admin.GET("/problem/:id/test_case/:test_case_id/output_file",
		controller.GetTestCaseOutputFile,
		middleware.HasPermission("read_problem_secret", "problem")).Name = "problem.getTestCaseOutputFile"

	admin.GET("/logs",
		controller.AdminGetLogs, middleware.HasPermission("read_logs")).Name = "admin.getLogs"

	if config.MustGet("debug", false).Value().(bool) {
		log.Debugf("Adding pprof handlers. SHOULD NOT BE USED IN PRODUCTION")
		e.Any("/debug/pprof/", func(c echo.Context) error {
			pprof.Index(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/*", func(c echo.Context) error {
			pprof.Index(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/cmdline", func(c echo.Context) error {
			pprof.Cmdline(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/profile", func(c echo.Context) error {
			pprof.Profile(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/symbol", func(c echo.Context) error {
			pprof.Symbol(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/trace", func(c echo.Context) error {
			pprof.Trace(c.Response().Writer, c.Request())
			return nil
		})
	}
}
